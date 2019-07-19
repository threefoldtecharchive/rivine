package config

import (
	"os"
	"path"
	"testing"
)

// TestGenerateConfigFileAndCheckFileExists creates a config file using 3 different file types.
// Check if the file is created properly
func TestGenerateConfigFileAndCheckFileExists(t *testing.T) {
	for _, typ := range []string{"toml", "yaml", "json"} {
		err := GenerateConfigFile(os.TempDir(), typ)
		if err != nil {
			t.Errorf("Error occured: %s", err)
		}

		_, err = os.Stat(path.Join(os.TempDir(), "blockchaincfg."+typ))
		if os.IsNotExist(err) {
			t.Errorf("File is not created, %s", err)
		}
		if err != nil {
			t.Errorf("Error occured: %s", err)
		}
		err = os.Remove(path.Join(os.TempDir(), "blockchaincfg."+typ))
		if err != nil {
			t.Errorf("Error occured removing file: %s", err)
		}
	}
}

// TestGenerateConfigFileWithUnknownFileType tries to create a config file with an unsupported filetype
// Should throw an error
func TestGenerateConfigFileWithUnknownFileType(t *testing.T) {
	typ := "test"
	err := GenerateConfigFile(os.TempDir(), typ)
	expectedError := "Filetype not supported"
	if err.Error() != expectedError {
		t.Errorf("Error actual: %s - and error expected: %s", err, expectedError)
	}
}
