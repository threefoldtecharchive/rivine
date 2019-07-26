package config

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const rootGithubAPIurl = "https://api.github.com"

func getTemplateRepo(owner, repo, version, destination string) (string, error) {
	endPoint := rootGithubAPIurl + path.Join("/repos", owner, repo, "tarball", version)
	fmt.Printf("Fetching repository: %s ...\n", endPoint)
	resp, err := http.Get(endPoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	err = untar(destination, resp.Body)
	if err != nil {
		return "", err
	}
	// Extract the commitHash from the headers, need this to rename directory later
	commitHash := strings.Split(strings.Split(strings.Split(resp.Header["Content-Disposition"][0], "=")[1], ".")[0], "-")[4]
	return commitHash, nil
}

func generateBlockchainTemplate(destinationDirPath, commitHash string, config *Config) error {
	// Directory where the contents of template repo is unpackged
	dirPath := destinationDirPath + "-" + commitHash

	// Remove everything in the destination directory path if it exists, since it would only contain what we generate
	err := os.RemoveAll(destinationDirPath)
	if err != nil {
		return err
	}
	// If destination Directory is not removed and it exists a rename will fail
	err = os.Rename(dirPath, destinationDirPath)
	if err != nil {
		return err
	}

	// Open the destination directory with our template contents
	f, _ := os.Open(destinationDirPath)
	fis, _ := f.Readdir(-1)
	f.Close()

	// For every file in this directory create our template code
	for _, fi := range fis {
		filePath := path.Join(destinationDirPath, fi.Name())
		isTemplate := path.Ext(fi.Name()) == ".template"
		if isTemplate {
			// First read the template file as string in order to write our generated code to a new file
			templateText, err := readTemplateFileAsString(filePath)
			if err != nil {
				return err
			}
			err = writeTemplateToFile(templateText, filePath, fi.Name(), config)
			if err != nil {
				return err
			}
			// Remove template file and keep generated one
			err = os.Remove(filePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func readTemplateFileAsString(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeTemplateToFile(templateText, filepath, filename string, config *Config) error {
	// Create a new file where will store generated code of this file
	newFilePath := strings.TrimSuffix(filepath, path.Ext(filename))
	file, err := os.Create(newFilePath)
	if err != nil {
		return err
	}
	// Create a new template and parse our template text
	t := template.Must(template.New("template").Parse(templateText))
	// Execute this template, which will fill in all templated values read from config
	return t.ExecuteTemplate(file, "template", config)
}

// untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		fmt.Printf("Unpackaged in: %s\n", target)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
