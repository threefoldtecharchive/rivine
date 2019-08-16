package config

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/otiai10/copy"
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
	dirPath := path.Join(destinationDirPath, config.Template.Repository.Owner) + "-" + config.Template.Repository.Repo + "-" + commitHash

	err := copy.Copy(dirPath, destinationDirPath)
	if err != nil {
		return err
	}

	// Remove generated files in old path
	err = os.RemoveAll(dirPath)
	if err != nil {
		return err
	}

	err = writeTemplateValues(destinationDirPath, config)
	if err != nil {
		return err
	}

	err = renameClientAndDaemonFolders(destinationDirPath, config)
	if err != nil {
		return err
	}

	err = removeTodoFolder(destinationDirPath)
	if err != nil {
		return err
	}
	return nil
}

func writeTemplateValues(destinationDirPath string, config *Config) error {
	fmap := template.FuncMap{
		"formatConditionAsString": formatConditionAsString,
	}

	err := filepath.Walk(destinationDirPath,
		func(fPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// fmt.Println(fPath, info.Size())
			isTemplate := path.Ext(info.Name()) == ".template"
			if isTemplate {
				// First read the template file as string in order to write our generated code to a new file
				templateText, err := readTemplateFileAsString(fPath)
				if err != nil {
					return err
				}
				err = writeTemplateToFile(templateText, fPath, info.Name(), config, fmap)
				if err != nil {
					return err
				}
				// Remove template file and keep generated one
				err = os.Remove(fPath)
				if err != nil {
					return err
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func renameClientAndDaemonFolders(destinationDirPath string, config *Config) error {
	oldClientFolderPath := path.Join(destinationDirPath, "todo", "cmd", "UNDEFINED_CLIENT_NAME")
	newClientFolderPath := path.Join(destinationDirPath, "todo", "cmd", config.Blockchain.Binaries.Client)
	daemonFolderPath := path.Join(destinationDirPath, "todo", "cmd", "UNDEFINED_DAEMON_NAME")
	newDaemonFolderPath := path.Join(destinationDirPath, "todo", "cmd", config.Blockchain.Binaries.Daemon)
	err := os.Rename(oldClientFolderPath, newClientFolderPath)
	if err != nil {
		return err
	}
	err = os.Rename(daemonFolderPath, newDaemonFolderPath)
	if err != nil {
		return err
	}
	return nil
}

func removeTodoFolder(destinationDirPath string) error {
	todoPath := path.Join(destinationDirPath, "todo")
	err := copy.Copy(todoPath, destinationDirPath)
	if err != nil {
		return err
	}
	err = os.RemoveAll(todoPath)
	if err != nil {
		return err
	}
	return nil
}

func formatConditionAsString(c Condition) (string, error) {
	return fmt.Sprintf("%s", string(c.UnlockHash().String())), nil
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

func writeTemplateToFile(templateText, filepath, filename string, config *Config, fmap template.FuncMap) error {
	// Create a new file where will store generated code of this file
	newFilePath := strings.TrimSuffix(filepath, path.Ext(filename))
	file, err := os.Create(newFilePath)
	if err != nil {
		return err
	}
	// Create a new template and parse our template text
	t := template.Must(template.New("template").Funcs(fmap).Parse(templateText))
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
