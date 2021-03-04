package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/inconshreveable/go-update"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/AB0529/prompter"
)

// TODO: Eventually have these links be auto-pulled
var (
	geckoLinux   = "https://github.com/mozilla/geckodriver/releases/download/v0.28.0/geckodriver-v0.28.0-linux64.tar.gz"
	geckoDarwin  = "https://github.com/mozilla/geckodriver/releases/download/v0.28.0/geckodriver-v0.28.0-macos.tar.gz"
	geckoWindows = "https://github.com/mozilla/geckodriver/releases/download/v0.28.0/geckodriver-v0.28.0-win64.zip"

	// Vendors path to the Vendors dir
	Vendors = "./Vendors"
	// Geckodriver Path to Gecko drivers
	Geckodriver = ""
	// Firefox Path to Firefox
	Firefox = ""
)

// IsDirEmpty determines if a directory is empty
func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.Readdir(1)

	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// DownloadFile will download a file off a link
func DownloadFile(dest string, url *url.URL) error {
	Info("Downloading for platform " + prompter.Yellow.Sprint(strings.ToUpper(runtime.GOOS)))

	// Get the file data
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write date to file
	_, err = io.Copy(f, resp.Body)
	return err
}

// Unzip will unzip a .zip file
func Unzip(src string, dest string) error {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		filePath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", filePath)
		}

		filenames = append(filenames, filePath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// Untar will untar a tarball
func Untar(src string, dest string) error {
	r, err := os.Open(src)
	defer r.Close()
	if err != nil {
		Error("Error while opening tarball")
		os.Exit(1)
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {

		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}
}

// DownloadAndUnzip will download a file and extract a zip
func DownloadAndUnzip(urlStr string, archiveStr string, dest string) {
	// Download the file
	requestURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		Error("Invalid URL!")
		os.Exit(1)
	}

	if err = DownloadFile(archiveStr, requestURL); err != nil {
		Error("Error while downloading")
		os.Exit(1)
	}

	// Unzip the file
	if err = Unzip(archiveStr, dest); err != nil {
		Error("Error while unzipping")
		os.Exit(1)
	}
}

// DownloadAndUntar will download a file and extract a tarball
func DownloadAndUntar(urlStr string, archiveStr string, dest string) {
	// Download the file
	requestURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		Error("Invalid URL!")
		os.Exit(1)
	}

	if err = DownloadFile(archiveStr, requestURL); err != nil {
		Error("Error while downloading ", err)
		os.Exit(1)
	}

	// Unzip the file
	if err = Untar(archiveStr, dest); err != nil {
		Error("Error while unzipping")
		os.Exit(1)
	}
}

// FindNewBinary will perform a GET request to find the latest binary url
func FindNewBinary() (GHResp, error) {
	var resp GHResp

	r, err := http.Get("https://api.github.com/repos/AB0529/lazy-meets/releases/latest")
	if err != nil {
		return GHResp{}, err
	}
	defer r.Body.Close()

	// Unmarshal the response to GHResp type
	reader, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(reader, &resp)
	if err != nil {
		return GHResp{}, err
	}

	return resp, nil
}

// doUpdate performs the update from the url
func doUpdate(resp GHResp) error {
	// Find correct binary for platform
	var githubURL string
	for _, assets := range resp.Assets {
		a := strings.Split(assets.Name, "-")
		binaryPlatform := strings.Replace(a[len(a) - 1], ".exe", "", 1)

		if binaryPlatform == runtime.GOOS {
			githubURL = assets.BrowserDownloadURL
			break
		}
	}

	// Perform update with url
	r, err := http.Get(githubURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	err = update.Apply(r.Body, update.Options{})

	Info("Done, you are now on version " + prompter.Green.Sprint(resp.TagName) + " please restart program!")

	return err
}

// Update will download latest Gecko drivers and Firefox and update from latest GitHub release
func Update() {
	// Make sure Vendors directory exists, if not create it
	if _, err := os.Stat(Vendors); os.IsNotExist(err) {
		os.Mkdir(Vendors, 0775)
	}

	empty, err := IsDirEmpty(Vendors)
	if err != nil {
		panic(err)
	}

	// Download the Gecko drivers for the specific platform
	switch runtime.GOOS {
	case "windows":
		if empty {
			DownloadAndUnzip(geckoWindows, Vendors+"/gecko.zip", Vendors)
		}

		Geckodriver = Vendors + "/geckodriver.exe"
		// Check if Firefox is installed
		_, err := os.Stat("C:\\Program Files\\Mozilla Firefox\\Firefox.exe")
		if err != nil {
			Error(prompter.Yellow.Sprint("Firefox not found ") + "make sure your Firefox installation is located at " + prompter.Red.Sprint("C:\\Program Files\\Mozilla Firefox\\Firefox.exe"))
			os.Exit(1)
		}
		Firefox = "C:\\Program Files\\Mozilla Firefox\\Firefox.exe"
		break
	case "darwin":
		if empty {
			DownloadAndUntar(geckoDarwin, Vendors+"/gecko.tar.gz", Vendors)
		}

		Geckodriver = Vendors + "/geckodriver"
		// Check if Firefox is installed
		_, err := os.Stat("/Applications/Firefox.app")
		if err != nil {
			Error(prompter.Yellow.Sprint("Firefox not found ") + "make sure your Firefox installation is located at " + prompter.Red.Sprint("/Applications/Firefox.app"))
			os.Exit(1)
		}
		Firefox = "/Applications/Firefox.app"
		break
	case "linux":
		if empty {
			DownloadAndUntar(geckoLinux, Vendors+"/gecko.tar.gz", Vendors)
		}
		// Makes sure Geckodriver is not running
		exec.Command("pkill geckodriver").Run()

		Geckodriver = Vendors + "/geckodriver"
		// Check if Firefox is installed
		_, err := os.Stat("/usr/bin/firefox")
		if err != nil {
			Error(prompter.Yellow.Sprint("Firefox not found ") + "make sure your Firefox installation is located at " + prompter.Red.Sprint("/usr/bin/firefox"))
			os.Exit(1)
		}
		Firefox = "/usr/bin/firefox"
		break
	default:
		Error(prompter.Yellow.Sprint(runtime.GOOS) + " is not a supported platform")
		os.Exit(1)
	}

	// Update binary file
	resp, err := FindNewBinary()
	if err != nil {
		panic(err)
	}

	// First time run
	if _, err := os.Stat(".update_cache"); err != nil {
		// Cache version number for first time run
		err = ioutil.WriteFile(".update_cache", []byte(resp.TagName), 0666)
		if err != nil {
			Error("error writing cache file")
			os.Exit(1)
		}

		// Perform the update
		err = doUpdate(resp)
		if err != nil {
			Error("error updating")
			os.Exit(1)
		}
	}

	// Check cache if version is the same or not
	dat, err := ioutil.ReadFile(".update_cache")
	if err != nil {
		Error("error reading cache file")
		os.Exit(1)
	}

	if string(dat) != resp.TagName {
		// New update found
		a := UpdateQuestion()
		// User says yes to update
		if a {
			// Perform the update
			err = doUpdate(resp)
			if err != nil {
				Error("error updating")
				os.Exit(1)
			}
			// Update cache file
			err = ioutil.WriteFile(".update_cache", []byte(resp.TagName), 0666)
			if err != nil {
				Error("error writing cache file")
				os.Exit(1)
			}
			time.Sleep(2*time.Second)
			os.Exit(1)
		}

		Info("Version " + prompter.Red.Sprint(string(dat)) + " is out of date!")
		return
	}

	Info("Version " + prompter.Green.Sprint(resp.TagName) + " is up to date!")
}
