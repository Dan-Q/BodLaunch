package main

import "os"
import "os/exec"
import "fmt"
import "io"
import "io/ioutil"
import "net/http"
import "strings"
import "archive/zip"
import "path/filepath"

func file_exists(path string) (bool) {
  _, err := os.Stat(path)
  if err == nil { return true }
  if os.IsNotExist(err) { return false }
  return true
}

// Unzip will un-compress a zip archive,
// moving all files and folders to an output directory
func unzip(src, dest string) ([]string, error) {
  var filenames []string
  r, err := zip.OpenReader(src)
  if err != nil {
      return filenames, err
  }
  defer r.Close()
  for _, f := range r.File {
    rc, err := f.Open()
    if err != nil {
      return filenames, err
    }
    defer rc.Close()
    // Store filename/path for returning and using later on
    fpath := filepath.Join(dest, f.Name)
    filenames = append(filenames, fpath)
    if f.FileInfo().IsDir() {
      // Make Folder
      os.MkdirAll(fpath, os.ModePerm)
    } else {
      // Make File
      var fdir string
      if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
        fdir = fpath[:lastIndex]
      }
      err = os.MkdirAll(fdir, os.ModePerm)
      if err != nil {
        return filenames, err
      }
      f, err := os.OpenFile(
          fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
      if err != nil {
        return filenames, err
      }
      defer f.Close()
      _, err = io.Copy(f, rc)
      if err != nil {
        return filenames, err
      }
    }
  }
  return filenames, nil
}

func main() {
  args := os.Args[1:]
  name, version_url, download_url, exe := string(args[0]), string(args[1]), string(args[2]), string(args[3])
  version_file, my_version := name + "/version", "0"
  // What are we doing
  fmt.Println("Preparing to launch " + name)
  // Determine current (my) version
  if(file_exists(name) && file_exists(version_file)) {
    loaded_my_version, _ := ioutil.ReadFile(version_file)
    my_version = strings.TrimSpace(string(loaded_my_version))
  }
  // Determine latest version
  remote_version_document, _ := http.Get(version_url)
  loaded_remote_version, _ := ioutil.ReadAll(remote_version_document.Body)
  remote_version := strings.TrimSpace(string(loaded_remote_version))
  // Download and extract latest version, if necessary
  if(remote_version != my_version){
    fmt.Println("Update required (" + my_version + " < " + remote_version + ")")
    // Download
    fmt.Println("Downloading update...")
    temp_file_name := name + "-tmp.zip"
    temp_file, _ := os.Create(temp_file_name)
    defer temp_file.Close()
    updated_version, _ := http.Get(download_url)
    defer updated_version.Body.Close()
    io.Copy(temp_file, updated_version.Body)
    updated_version.Body.Close()
    temp_file.Close()
    // Unzip
    fmt.Println("Unpacking update...")
    unzip(temp_file_name, "./")
    // Delete temporary files
    os.Remove(temp_file_name)
  }
  // Run!
  fmt.Println("Launching " + name + "...")
  // TODO
  os.Chdir(name)
  cmd := exec.Command(exe)
  cmd.Run()
}
