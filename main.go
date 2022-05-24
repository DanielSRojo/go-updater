package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {

	// Get latest version of Go from go.dev/dl
	goLatestVer, err := getGoLatestVersion()
	if err != nil {
		fmt.Println("error: couldn't get golang version from https://go.dev/dl/")
		return
	}
	fmt.Printf("Latest version found at go.dev/dl: %s\n", goLatestVer)
	goLatestVer = strings.Split(goLatestVer, "go")[1]
	downloadUrl := fmt.Sprintf("https://go.dev/dl/go%s.linux-amd64.tar.gz", goLatestVer)

	// Get current installed version
	goCurrentVer, err := getInstalledVersion()
	if err != nil {
		goCurrentVer = "go0.0.0"
		fmt.Println("Go installation not found on this system. Installing it now...")
	} else {
		fmt.Printf("Installed version on this system: %s\n", goCurrentVer)
	}
	goCurrentVer = strings.Split(goCurrentVer, "go")[1]

	// Reformat installed and latest go versions to []uint64 type
	var current, target []uint64
	for _, v := range strings.Split(goCurrentVer, ".") {
		n, err := strconv.ParseUint(v, 0, 64)
		if err != nil {
			fmt.Printf("error: couldn't convert %s to int", v)
			return
		}
		current = append(current, n)
	}
	for _, v := range strings.Split(goLatestVer, ".") {
		n, err := strconv.ParseUint(v, 0, 64)
		if err != nil {
			fmt.Printf("error: couldn't convert %s to int", v)
			return
		}
		target = append(target, n)
	}

	// Check if latest version is newer than current, exit otherwise
	if !isNewerVersion(current, target) {
		fmt.Println("Upgrade not needed")
		return
	}

	// Download package from source
	goLatestVer = fmt.Sprintf("go%s", goLatestVer)

	packagePath := fmt.Sprintf("/tmp/%s.linux-amd64.tar.gz", goLatestVer)
	_, err = os.Stat(packagePath)
	if os.IsNotExist(err) {
		fmt.Printf("Downloading %s from %s\n", goLatestVer, downloadUrl)
		err = getPackage(goLatestVer, packagePath);
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Uncompress downloaded package
	tarFile := strings.Replace(packagePath, ".gz", "", 1)
	if err = unGzip(packagePath, tarFile); err != nil {
		fmt.Println(err)
		return
	}

	// Remove previous installation
	goInstallDir := path.Join("/", "usr", "local")
	if err = os.RemoveAll(goInstallDir); err != nil {
		fmt.Println(err)
		return
	}

	// Unarchive package and set to installation directory
	err = unTar(tarFile, goInstallDir)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Success, golang version %s installed correctly\n", goLatestVer)
}

// getGoLatestVersion gets golang linux-amd64 latest version from go.dev
func getGoLatestVersion() (string, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://go.dev/dl/", nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	s := bufio.NewScanner(res.Body)
	var version string
	for s.Scan() {
		if strings.Contains(s.Text(), "span class") && strings.Contains(s.Text(), "linux-amd64") {
			version = strings.Split(s.Text(), ">")[1]
			version = strings.Split(version, ".linux-amd64.tar.gz")[0]
			break
		}
	}

	return version, err
}

// getPackage downloads golang linux-amd64 given version from go.dev/dl/
// and writes it into /tmp. If success, it returns downloaded file path.
func getPackage(version, target string) (error) {
	path := fmt.Sprintf("/tmp/%s.linux-amd64.tar.gz", version)
	url := fmt.Sprintf("https://go.dev/dl/%s.linux-amd64.tar.gz", version)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("error: invalid response, status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, body, 0664)
	if err != nil {
		return fmt.Errorf("error: couldn't write file %s", path)
	}

	return nil
}

// isNewerVersion compares two given arrays and returns true if the first one
// is newest than the second one. Format used is mayor.minor.patch as []uint64
func isNewerVersion(current, target []uint64) bool {

	if target[0] > current[0] {
		return true
	} else if target[0] == current[0] && target[1] > current[1] {
		return true
	} else if target[1] == current[1] && target[2] > current[2] {
		return true
	}

	return false
}

func unGzip(source, target string) error {

	r, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("error while opening %s", source)
	}
	defer r.Close()

	archive, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	w, err := os.Create(target)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error writing on %s", target)
	}
	defer w.Close()

	_, err = io.Copy(w, archive)
	return err
}

func unTar(source, target string) error {
	r, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("error: couldn't open file %s", source)
	}
	defer r.Close()

	rTar := tar.NewReader(r)

	for {
		h, err := rTar.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("error: couldn't read %s", source)
		}

		path := filepath.Join(target, h.Name)
		info := h.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return fmt.Errorf("error:  couldn't write file %s", path)
			}
			continue
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return fmt.Errorf("error: couldn't open file %s", path)
		}
		defer f.Close()

		_, err = io.Copy(f, rTar)
		if err != nil {
			return fmt.Errorf("error: couldn't write on file %s", path)
		}
	}

	return nil
}

func getInstalledVersion() (string, error) {
	path :="/usr/local/go/VERSION"
	_, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("error: go installation not found")
	}

	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("error: couldn't open file %s", path)
	}
	defer f.Close()

	version, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error: couldn't read file %s", path)
	}

	return string(version), nil
}