package persist

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/types"
)

// TestLogger checks that the basic functions of the file logger work as
// designed.
func TestLogger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Create a folder for the log file.
	testdir := build.TempDir(persistDir, t.Name())
	err := os.MkdirAll(testdir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Create the logger.
	logFilename := filepath.Join(testdir, "test.log")
	fl, err := NewFileLogger(types.DefaultBlockchainInfo(), logFilename, false)
	if err != nil {
		t.Fatal(err)
	}

	// Write an example statement, and then close the logger.
	fl.Println("TEST: this should get written to the logfile")
	err = fl.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Check that data was written to the log file. There should be three
	// lines, one for startup, the example line, and one to close the logger.
	expectedSubstring := []string{"STARTUP", "TEST", "SHUTDOWN", ""} // file ends with a newline
	validatelogfile(t, logFilename, expectedSubstring, 4)

}

// TestLoggerCritical prints a critical message from the logger.
func TestLoggerCritical(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Create a folder for the log file.
	testdir := build.TempDir(persistDir, t.Name())
	err := os.MkdirAll(testdir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Create the logger.
	logFilename := filepath.Join(testdir, "test.log")
	fl, err := NewFileLogger(types.DefaultBlockchainInfo(), logFilename, false)
	if err != nil {
		t.Fatal(err)
	}

	// Write a catch for a panic that should trigger when logger.Critical is
	// called.
	defer func() {
		r := recover()
		if r == nil {
			t.Error("critical message was not thrown in a panic")
		}

		// Close the file logger to clean up the test.
		err = fl.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	fl.Critical("a critical message")
}

func TestVerboseLogger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Create a folder for the log file.
	testdir := build.TempDir(persistDir, t.Name())
	err := os.MkdirAll(testdir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Create the logger.
	logFilename := filepath.Join(testdir, "test.log")
	fl, err := NewFileLogger(types.DefaultBlockchainInfo(), logFilename, true)
	if err != nil {
		t.Fatal(err)
	}

	// Write an example statement, and then close the logger.
	fl.Debugln("ROBTEST: this should get written to the logfile")
	err = fl.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Check that data was written to the log file. There should be three
	// lines, one for startup, the example line, and one to close the logger.
	expectedSubstring := []string{"STARTUP", "ROBTEST", "SHUTDOWN", ""} // file ends with a newline
	validatelogfile(t, logFilename, expectedSubstring, 4)

	// Create the logger.
	logFilename = filepath.Join(testdir, "test.log2")
	fl, err = NewFileLogger(types.DefaultBlockchainInfo(), logFilename, false)
	if err != nil {
		t.Fatal(err)
	}

	// Write an example statement, and then close the logger.
	fl.Debugln("ROBTEST: this should not get written to the logfile")
	err = fl.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Check that data was written to the log file. There should be three
	// lines, one for startup, the example line, and one to close the logger.
	expectedSubstring = []string{"STARTUP", "SHUTDOWN", ""} // file ends with a newline
	validatelogfile(t, logFilename, expectedSubstring, 3)

}
func validatelogfile(t *testing.T, logFilename string, expectedSubstrings []string, numberOfLines int) {
	fileData, err := ioutil.ReadFile(logFilename)
	if err != nil {
		t.Fatal(err)
	}
	fileLines := strings.Split(string(fileData), "\n")
	for i, line := range fileLines {
		if !strings.Contains(string(line), expectedSubstrings[i]) {
			t.Error("did not find the expected message in the logger")
		}
	}
	if len(fileLines) != numberOfLines {
		t.Error("logger did not create the correct number of lines:", len(fileLines))
	}
}
