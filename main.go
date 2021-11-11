package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

const REGISTRY_BASE = "https://registry.npmjs.org"

type PackageDist struct {
	TarballUrl string `json:"tarball"`
	ShaSum 	 string `json:"shasum"`
}

type PackageVersion struct {
	Name string `json:"name"`
	Version string `json:"version"`
	Dist PackageDist `json:"dist"`
}

type PackageInfo struct {
	Name string `json:"name"`
	Description string `json:"description"`
	DistTags map[string]string `json:"dist-tags"`
	Versions map[string]PackageVersion `json:"versions"`
}

func GetPackageInfo(name string) (PackageInfo, error) {
	packageInfo := PackageInfo{}
	response, err := http.Get(REGISTRY_BASE + "/" + name)
	if err != nil {
		return packageInfo, err
	}

	// Read response to json
	err = json.NewDecoder(response.Body).Decode(&packageInfo)
	if err != nil {
		return packageInfo, err
	}

	return packageInfo, nil
}

func DownloadTarball(url string, outputPath string) error {
	respone, err := http.Get(url)
	if err != nil {
		return nil
	}

	defer respone.Body.Close()

	archive, err := gzip.NewReader(respone.Body)
	if err != nil {
		return nil
	}

	tarReader := tar.NewReader(archive)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// Change path to same name but up a directory.

		outLocation := regexp.MustCompile(`^[^\/]+`).ReplaceAllString(header.Name, "")

		path := filepath.Join(outputPath, outLocation)
		info := header.FileInfo()
		
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		// Create directory if it doesn't exist.
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(path, os.O_CREATE | os.O_TRUNC | os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	packageName := "express"

	packageInfo, err := GetPackageInfo(packageName)
	if err != nil {
		fmt.Println(err)
		return
	}

	latestVersion := packageInfo.DistTags["latest"]
	latestPackage := packageInfo.Versions[latestVersion]

	tarballUrl := latestPackage.Dist.TarballUrl

	err = os.MkdirAll("./testing", 0755)


	err = DownloadTarball(tarballUrl, "./testing")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(tarballUrl)
}