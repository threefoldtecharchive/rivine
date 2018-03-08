package datastore

import (
	"testing"

	"github.com/rivine/rivine/types"
)

func TestSerialize(t *testing.T) {
	nsm := &NamespaceManager{State: NamespaceManagerState{BlockHeight: types.BlockHeight(5)}}
	nsm2 := &NamespaceManager{}

	amBytes, err := nsm.Serialize()
	if err != nil {
		t.Fatal("Serializing namespace manager failed: ", err)
	}
	err = nsm2.Deserialize(amBytes)
	if err != nil {
		t.Fatal("Deserializing namespace manager failed: ", err)
	}
	if nsm.State != nsm2.State {
		t.Fatal("Deserialized state does not match: ", nsm2.State.BlockHeight)
	}
}
