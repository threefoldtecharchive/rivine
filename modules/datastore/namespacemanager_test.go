package datastore

import (
	"testing"

	"github.com/threefoldtech/rivine/types"
)

func TestSerialize(t *testing.T) {
	nsm := &namespaceManager{state: namespaceManagerState{BlockHeight: types.BlockHeight(5)}}
	nsm2 := &namespaceManager{}

	amBytes, err := nsm.serialize()
	if err != nil {
		t.Fatal("Serializing namespace manager failed: ", err)
	}
	err = nsm2.deserialize(amBytes)
	if err != nil {
		t.Fatal("Deserializing namespace manager failed: ", err)
	}
	if nsm.state != nsm2.state {
		t.Fatal("Deserialized state does not match: ", nsm2.state.BlockHeight)
	}
}
