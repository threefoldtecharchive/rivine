package config

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/otiai10/copy"

	"github.com/threefoldtech/rivine/types"
)

const rootGithubAPIurl = "https://api.github.com"

var _githubRgxp = regexp.MustCompile(`github.com/([^/]+)/([^/]+)`)

func githubOwnerAndRepoFromString(s string) (string, string, error) {
	repoGithubMatches := _githubRgxp.FindStringSubmatch(s)
	if len(repoGithubMatches) != 3 {
		return "", "", fmt.Errorf("invalid repository (only Github is supported at the moment): %s", s)
	}
	return repoGithubMatches[1], repoGithubMatches[2], nil
}

// getTemplateRepo fetches the template repository from github and extracts this tar file.
// At the end of this function we extract the commithash from the headers in order to rename this extracted directory later.
func getTemplateRepo(repository, version, destination string) (string, error) {
	templOwner, templRepo, err := githubOwnerAndRepoFromString(repository)
	if err != nil {
		return "", err
	}

	endPoint := rootGithubAPIurl + path.Join("/repos", templOwner, templRepo, "tarball", version)
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

type TemplateConfig struct {
	Frontend TemplateFrontendConfig `json:"frontend" validate:"required"`
}

type TemplateFrontendConfig struct {
	Explorer map[string]TemplateFrontendTypeConfig `json:"explorer" validate:"required"`
	Faucet   map[string]TemplateFrontendTypeConfig `json:"Faucet" validate:"required"`
}

type TemplateFrontendTypeConfig struct {
	Repository string `json:"repo" validate:"required"`
	Branch     string `json:"branch" validate:"required"`
}

func generateBlockchainTemplate(destinationDirPath, commitHash string, config *Config, opts *BlockchainGenerationOpts) error {
	var templateConfig *TemplateConfig

	var fPathAction func(fPath, dirPath, destPath string) error
	if config.Generation != nil && len(config.Generation.Ignore) > 0 {
		fPathAction = func(fPath, dirPath, destPath string) error {
			relFilePath := strings.TrimLeft(strings.TrimPrefix(fPath, dirPath), `\/`)
			cleanRelFilePath := strings.TrimSuffix(relFilePath, ".template")
			for _, p := range config.Generation.Ignore {
				if p.Match(cleanRelFilePath) {
					return nil
				}
			}
			return copy.Copy(fPath, path.Join(destPath, relFilePath))
		}
	} else {
		fPathAction = func(fPath, dirPath, destPath string) error {
			relFilePath := strings.TrimLeft(strings.TrimPrefix(fPath, dirPath), `\/`)
			return copy.Copy(fPath, path.Join(destPath, relFilePath))
		}
	}

	templOwner, templRepo, err := githubOwnerAndRepoFromString(config.Template.Repository)
	if err != nil {
		return err
	}
	// Directory where the contents of template repo is unpackged
	dirPath := path.Join(destinationDirPath, templOwner+"-"+templRepo+"-"+commitHash)

	// walk over the files, and copy only those not ignored
	err = filepath.Walk(dirPath, func(fPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err // return an error immediately
		}

		// ignore directories
		if info.IsDir() {
			return nil
		}

		// consume template.json config file instead of copying it
		if fPath == path.Join(dirPath, "template.json") {
			templateJSONFile, err := os.Open(fPath)
			if err != nil {
				return fmt.Errorf("failed to open special template config file: %v", err)
			}
			templateConfig = new(TemplateConfig)
			err = json.NewDecoder(templateJSONFile).Decode(templateConfig)
			if err != nil {
				return fmt.Errorf("failed to decode special template config file: %v", err)
			}
			return validate.Struct(templateConfig) // finished here, no need to copy it
		}

		return fPathAction(fPath, dirPath, destinationDirPath)
	})
	if err != nil {
		return err
	}

	// Remove generated files in old path
	err = os.RemoveAll(dirPath)
	if err != nil {
		return err
	}

	// generate optionally also explorer frontend (opt-out)
	if opts != nil && opts.FrontendExplorerType != FrontendExplorerTypeNone {
		if templateConfig == nil {
			return errors.New("an explorer frontend type is selected but usen template repo doesn't link to any explorer frontend template")
		}
		frontendExplorerTypeStr := opts.FrontendExplorerType.String()
		explorerFrontendConfig, ok := templateConfig.Frontend.Explorer[frontendExplorerTypeStr]
		if !ok {
			return fmt.Errorf("used template repo doesn't link to an explorer frontend template of type %s (%d)", frontendExplorerTypeStr, opts.FrontendExplorerType)
		}

		explorerTemplOwner, explorerTemplRepo, err := githubOwnerAndRepoFromString(explorerFrontendConfig.Repository)
		if err != nil {
			return fmt.Errorf("invalid frontend explorer (type: %s) template repo %s: %v", frontendExplorerTypeStr, explorerFrontendConfig.Repository, err)
		}
		commitHash, err := getTemplateRepo(explorerFrontendConfig.Repository, explorerFrontendConfig.Branch, destinationDirPath)
		if err != nil {
			return fmt.Errorf("failed to download frontend explorer (type: %s) template repo %s: %v", frontendExplorerTypeStr, explorerFrontendConfig.Repository, err)
		}

		// Directory where the contents of template repo is unpacked
		frontendExplorerDirPath := path.Join(destinationDirPath, explorerTemplOwner+"-"+explorerTemplRepo+"-"+commitHash)

		// Directory where the frontend explorer needs to be generated to
		frontendExplorerDestinationPath := path.Join(destinationDirPath, "frontend", "explorer")

		// modify fPathAction with ignore option here,
		// as we need to ensure that the frontend/explorer path is prefixed
		fPathAction := fPathAction
		if config.Generation != nil && len(config.Generation.Ignore) > 0 {
			fPathAction = func(fPath, dirPath, destPath string) error {
				relFilePath := strings.TrimLeft(strings.TrimPrefix(fPath, dirPath), `\/`)
				cleanRelFilePath := path.Join("frontend", "explorer", strings.TrimSuffix(relFilePath, ".template"))
				for _, p := range config.Generation.Ignore {
					if p.Match(cleanRelFilePath) {
						return nil
					}
				}
				return copy.Copy(fPath, path.Join(destPath, relFilePath))
			}
		}

		// walk over the files, and copy only those not ignored
		err = filepath.Walk(frontendExplorerDirPath, func(fPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err // return an error immediately
			}

			// ignore directories
			if info.IsDir() {
				return nil
			}

			return fPathAction(fPath, frontendExplorerDirPath, frontendExplorerDestinationPath)
		})
		if err != nil {
			return err
		}

		// Remove generated files in old path
		err = os.RemoveAll(frontendExplorerDirPath)
		if err != nil {
			return err
		}
	}

	// generate faucet by default (opt-in)
	if opts.FrontendFaucet != true {
		if templateConfig == nil {
			return errors.New("an explorer frontend type is selected but usen template repo doesn't link to any explorer frontend template")
		}
		explorerFaucetConfig, ok := templateConfig.Frontend.Faucet["go"]
		if !ok {
			return fmt.Errorf("used template repo doesn't link to an faucet frontend template of type go")
		}

		faucetTemplOwner, faucetTempRepo, err := githubOwnerAndRepoFromString(explorerFaucetConfig.Repository)
		if err != nil {
			return fmt.Errorf("invalid faucet explorer (type: go) template repo %s: %v", explorerFaucetConfig.Repository, err)
		}
		commitHash, err := getTemplateRepo(explorerFaucetConfig.Repository, explorerFaucetConfig.Branch, destinationDirPath)
		if err != nil {
			return fmt.Errorf("failed to download faucet explorer (type: go) template repo %s: %v", explorerFaucetConfig.Repository, err)
		}

		// Directory where the contents of template repo is unpacked
		frontendExplorerDirPath := path.Join(destinationDirPath, faucetTemplOwner+"-"+faucetTempRepo+"-"+commitHash)

		// Directory where the frontend explorer needs to be generated to
		frontendExplorerDestinationPath := path.Join(destinationDirPath, "frontend", "faucet")

		// modify fPathAction with ignore option here,
		// as we need to ensure that the frontend/explorer path is prefixed
		fPathAction := fPathAction
		if config.Generation != nil && len(config.Generation.Ignore) > 0 {
			fPathAction = func(fPath, dirPath, destPath string) error {
				relFilePath := strings.TrimLeft(strings.TrimPrefix(fPath, dirPath), `\/`)
				cleanRelFilePath := path.Join("frontend", "faucet", strings.TrimSuffix(relFilePath, ".template"))
				for _, p := range config.Generation.Ignore {
					if p.Match(cleanRelFilePath) {
						return nil
					}
				}
				return copy.Copy(fPath, path.Join(destPath, relFilePath))
			}
		}

		// walk over the files, and copy only those not ignored
		err = filepath.Walk(frontendExplorerDirPath, func(fPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err // return an error immediately
			}

			// ignore directories
			if info.IsDir() {
				return nil
			}

			return fPathAction(fPath, frontendExplorerDirPath, frontendExplorerDestinationPath)
		})
		if err != nil {
			return err
		}

		// Remove generated files in old path
		err = os.RemoveAll(frontendExplorerDirPath)
		if err != nil {
			return err
		}
	}

	err = writeTemplateValues(destinationDirPath, config)
	if err != nil {
		return err
	}

	err = renameClientAndDaemonFolders(destinationDirPath, config)
	if err != nil {
		return err
	}

	// optionally Go-Format the codebase (enabled by default)
	if config.Generation == nil || !config.Generation.DisableGoFormatting {
		goDirs := []string{
			path.Join(destinationDirPath, "cmd"),
			path.Join(destinationDirPath, "pkg"),
		}
		// default Go Imports
		err = exec.Command("goimports", append([]string{"-w"}, goDirs...)...).Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to exec goimports: %v\n", err)
		}
		// default Go Formatting
		err = exec.Command("gofmt", append([]string{"-s", "-w"}, goDirs...)...).Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to exec gofmt: %v\n", err)
		}
	}

	return nil
}

func writeTemplateValues(destinationDirPath string, config *Config) error {
	fmap := template.FuncMap{
		"formatConditionAsUnlockhashString":            formatConditionAsUnlockhashString,
		"formatConditionAsGoString":                    formatConditionAsGoString,
		"formatValueStringAsOneCoinCurrencyMultiplier": formatValueStringAsOneCoinCurrencyMultiplier,
	}
	for n, f := range sprig.FuncMap() {
		fmap[n] = f
	}

	err := filepath.Walk(destinationDirPath,
		func(fPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path.Ext(info.Name()) != ".template" {
				return nil
			}
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
			return os.Remove(fPath)
		})
	if err != nil {
		return err
	}
	return nil
}

func renameClientAndDaemonFolders(destinationDirPath string, config *Config) error {
	oldClientFolderPath := path.Join(destinationDirPath, "cmd", "UNDEFINED_CLIENT_NAME")
	newClientFolderPath := path.Join(destinationDirPath, "cmd", config.Blockchain.Binaries.Client)
	daemonFolderPath := path.Join(destinationDirPath, "cmd", "UNDEFINED_DAEMON_NAME")
	newDaemonFolderPath := path.Join(destinationDirPath, "cmd", config.Blockchain.Binaries.Daemon)
	if _, err := os.Stat(newClientFolderPath); !os.IsNotExist(err) {
		err = os.RemoveAll(newClientFolderPath)
		if err != nil {
			return fmt.Errorf("failed to delete existing client dir: %v", err)
		}
	}
	err := os.Rename(oldClientFolderPath, newClientFolderPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(newDaemonFolderPath); !os.IsNotExist(err) {
		err = os.RemoveAll(newDaemonFolderPath)
		if err != nil {
			return fmt.Errorf("failed to delete existing daemon dir: %v", err)
		}
	}
	err = os.Rename(daemonFolderPath, newDaemonFolderPath)
	if err != nil {
		return err
	}
	return nil
}

func formatConditionAsUnlockhashString(c Condition) (string, error) {
	ct := c.ConditionType()
	if ct == types.ConditionTypeTimeLock {
		c = Condition{types.NewCondition(c.Condition.(*types.TimeLockCondition))}
		ct = c.ConditionType()
	}
	if ct == types.ConditionTypeUnlockHash {
		return fmt.Sprintf("%s", string(c.UnlockHash().String())), nil
	}
	return "", fmt.Errorf("cannot marshal unsupported condition of type %d", ct)
}

func formatConditionAsGoString(c Condition) (string, error) {
	ct := c.ConditionType()
	if ct == types.ConditionTypeUnlockHash {
		return fmt.Sprintf(
			`types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("%s")))`,
			c.UnlockHash().String()), nil
	}
	if ct == types.ConditionTypeMultiSignature {
		msc := c.Condition.(*types.MultiSignatureCondition)
		// validate
		if len(msc.UnlockHashes) == 0 {
			return "", errors.New("MultiSig outputs must specify at least a single address which can sign it as an input")
		}
		if msc.MinimumSignatureCount == 0 {
			return "", errors.New("MultiSig outputs must specify amount of signatures required")
		}
		// return it as a golang string
		unlockhashes := make([]string, 0, len(msc.UnlockHashes))
		for _, uh := range msc.UnlockHashes {
			unlockhashes = append(unlockhashes, fmt.Sprintf(`unlockHashFromHex("%s")`, uh.String()))
		}
		return fmt.Sprintf(
			`types.NewCondition(types.NewMultiSignatureCondition(types.UnlockHashSlice{%s}, %d))`,
			strings.Join(unlockhashes, ", "), msc.MinimumSignatureCount), nil
	}
	return "", fmt.Errorf("cannot marshal unsupported condition of type %d", ct)
}

func formatValueStringAsOneCoinCurrencyMultiplier(v string) (string, error) {
	parts := strings.Split(v, ".")
	if len(parts) == 1 {
		// assume it is a natural number
		return fmt.Sprintf(".Mul64(%s)", parts[0]), nil
	}
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid one coin currency value %s", v)
	}
	parts[1] = strings.TrimRight(parts[1], "0")
	if parts[1] == "" {
		// assume it is a natural number
		return fmt.Sprintf(".Mul64(%s)", parts[0]), nil
	}
	// assume it is a real number
	parts[0] = strings.TrimLeft(parts[0], "0")
	if parts[0] == "" {
		if parts[1] == "1" {
			// real number that functions as a divisor only
			return fmt.Sprintf(".Div64(1%s)", strings.Repeat("0", len(parts[1]))), nil
		}
		// real number with decimals only
		return fmt.Sprintf(".Mul64(%s).Div64(1%s)", parts[1], strings.Repeat("0", len(parts[1]))), nil
	}
	// complete real number
	return fmt.Sprintf(".Mul64(%s%s).Div64(1%s)", parts[0], parts[1], strings.Repeat("0", len(parts[1]))), nil
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

		// if the header is nil, just skip it
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		fmt.Printf("Unpackaged in: %s\n", target)

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
