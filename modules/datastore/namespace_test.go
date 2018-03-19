package datastore

import (
	"testing"
)

var (
	testNameBytes = [NamespaceBytes]byte{'t', 'e', 's', 't'}
	testName      = string(testNameBytes[:])
)

func TestString(t *testing.T) {
	testNs := Namespace(testNameBytes)
	if nsString := testNs.String(); nsString != testName {
		t.Error("Expected namespace string to be \"test\", but got ", nsString)
	}
}

func TestLoadString(t *testing.T) {
	testNs := Namespace{}
	if err := testNs.LoadString(testName); err != nil {
		t.Error("Error while loading namespace string: ", err)
	}
	nsBytes := [NamespaceBytes]byte{}
	copy(nsBytes[:], testNs[:])
	if nsBytes != testNameBytes {
		t.Error("Namespace string is not loaded correctly")
	}

	if err := testNs.LoadString(testName + testName); err == nil {
		t.Error("No error while loading faulty namespace string")
	}
}
