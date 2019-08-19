package config

import (
	"os"
	"path"
	"testing"
)

// TestGenerateConfigFileWithUnknownFileType tries to create a config file with an unsupported filetype
// Should throw an error
func TestGenerateConfigFileWithUnknownFileType(t *testing.T) {
	typ := "test"
	filepath := path.Join(os.TempDir(), "blockchaincfg."+typ)
	err := GenerateConfigFile(filepath)
	expectedError := "Filetype not supported"
	if err != ErrUnsupportedFileType {
		t.Errorf("Error actual: %s - and error expected: %s", err, expectedError)
	}
}

// TestGenerateAndLoadConfigFile creates a default config file with some default values using 2 different file types.
// Check if the file is created properly and removes it after this test ends
// Then it loads this configuration, validates and fills in some default values if these are not provided.
func TestGenerateAndLoadConfigFile(t *testing.T) {
	for _, typ := range []string{"yaml", "json"} {
		filepath := path.Join(os.TempDir(), "blockchaincfg."+typ)
		err := GenerateConfigFile(filepath)
		if err != nil {
			t.Errorf("Error occured: %s", err)
		}

		file, err := os.Open(filepath)
		if err != nil {
			t.Errorf("Error occured: %s", err)
		}

		_, err = loadConfig("."+typ, file)
		if err != nil {
			t.Errorf("Error occured loading file: %s", err)
		}
		_, err = os.Stat(filepath)
		if os.IsNotExist(err) {
			t.Errorf("File is not created, %s", err)
		}
		if err != nil {
			t.Errorf("Error occured: %s", err)
		}
		err = os.Remove(filepath)
		if err != nil {
			t.Errorf("Error occured removing file: %s", err)
		}
	}
}

// TestGenerateAndLoadConfigFile creates a default config file with some default values using 2 different file types.
// Check if the file is created properly and removes it after this test ends
// Then it loads this configuration, validates and fills in some default values if these are not provided.
func TestGenerateAndLoadConfigFileAndGetTemplateRepoAndGenerateBlockchainCode(t *testing.T) {
	for _, typ := range []string{"yaml", "json"} {
		filepath := path.Join(os.TempDir(), "blockchaincfg."+typ)
		err := GenerateConfigFile(filepath)
		if err != nil {
			t.Errorf("Error occured: %s", err)
		}

		file, err := os.Open(filepath)
		if err != nil {
			t.Errorf("Error occured: %s", err)
		}

		conf, err := loadConfig("."+typ, file)
		if err != nil {
			t.Errorf("Error occured loading file: %s", err)
		}

		dirpath := path.Join(os.TempDir(), "tmpbctest")
		commitHash, err := getTemplateRepo(conf.Template.Repository.Owner, conf.Template.Repository.Repo, conf.Template.Version, dirpath)
		if err != nil {
			t.Errorf("Error occured fetching template repo: %s", err)
		}

		err = generateBlockchainTemplate(dirpath, commitHash, conf)
		if err != nil {
			t.Errorf("Error occured generating blockchain code: %s", err)
		}

		_, err = os.Stat(filepath)
		if os.IsNotExist(err) {
			t.Errorf("File is not created, %s", err)
		}
		if err != nil {
			t.Errorf("Error occured: %s", err)
		}
		err = os.Remove(filepath)
		if err != nil {
			t.Errorf("Error occured removing file: %s", err)
		}
	}
}
