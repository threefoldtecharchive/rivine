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
// func TestGenerateAndLoadConfigFile(t *testing.T) {
// 	for _, typ := range []string{"yaml", "json"} {
// 		filepath := path.Join(os.TempDir(), "blockchaincfg."+typ)
// 		err := GenerateConfigFile(filepath)
// 		if err != nil {
// 			t.Errorf("Error occured: %s", err)
// 		}

// 		err = LoadConfigFile(filepath)
// 		if err != nil {
// 			t.Errorf("Error occured loading file: %s", err)
// 		}
// 		_, err = os.Stat(filepath)
// 		if os.IsNotExist(err) {
// 			t.Errorf("File is not created, %s", err)
// 		}
// 		if err != nil {
// 			t.Errorf("Error occured: %s", err)
// 		}
// 		err = os.Remove(filepath)
// 		if err != nil {
// 			t.Errorf("Error occured removing file: %s", err)
// 		}
// 	}
// }
