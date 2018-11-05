package datastore

import (
	"reflect"
	"testing"

	"github.com/threefoldtech/rivine/types"
)

func Test_parseData(t *testing.T) {
	tests := []struct {
		name    string
		rawData []byte
		sp      types.Specifier
		ns      Namespace
		data    []byte
	}{
		{
			name:    "test-ns",
			rawData: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 88, 85, 79, 52, 4, 6, 8},
			sp:      types.Specifier{},
			ns:      Namespace{88, 85, 79, 52},
			data:    []byte{4, 6, 8},
		},
		{
			name:    "test-shortdata",
			rawData: []byte{146, 88, 46, 52, 79, 5},
			sp:      types.Specifier{},
			ns:      Namespace{},
			data:    nil,
		},
		{
			name:    "test-specifier",
			rawData: []byte{126, 84, 15, 79, 42, 79, 241, 1, 89, 190, 21, 78, 16, 17, 15, 89, 0, 0, 0, 0, 65, 1, 5, 61, 79, 8, 5, 4},
			sp:      types.Specifier{126, 84, 15, 79, 42, 79, 241, 1, 89, 190, 21, 78, 16, 17, 15, 89},
			ns:      Namespace{},
			data:    []byte{65, 1, 5, 61, 79, 8, 5, 4},
		},
		{
			name:    "test-data",
			rawData: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 54, 94, 78, 26, 78, 123, 24, 205, 192, 47, 53, 74, 2},
			sp:      types.Specifier{},
			ns:      Namespace{},
			data:    []byte{54, 94, 78, 26, 78, 123, 24, 205, 192, 47, 53, 74, 2},
		},
		{
			name:    "test-full",
			rawData: []byte{85, 44, 96, 6, 74, 3, 79, 108, 241, 100, 255, 0, 14, 28, 49, 53, 109, 58, 49, 204, 79, 10, 65, 71, 195, 0, 7, 14, 36, 8, 209, 44, 50, 54, 24, 0, 31, 80, 31, 46, 21, 100, 24, 83, 0, 0, 23, 4, 50},
			sp:      types.Specifier{85, 44, 96, 6, 74, 3, 79, 108, 241, 100, 255, 0, 14, 28, 49, 53},
			ns:      Namespace{109, 58, 49, 204},
			data:    []byte{79, 10, 65, 71, 195, 0, 7, 14, 36, 8, 209, 44, 50, 54, 24, 0, 31, 80, 31, 46, 21, 100, 24, 83, 0, 0, 23, 4, 50},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp, ns, data := parseData(tt.rawData)
			if !reflect.DeepEqual(sp, tt.sp) {
				t.Errorf("parseData() sp = %v, want %v", sp, tt.sp)
			}
			if !reflect.DeepEqual(ns, tt.ns) {
				t.Errorf("parseData() ns = %v, want %v", ns, tt.ns)
			}
			if !reflect.DeepEqual(data, tt.data) {
				t.Errorf("parseData() data = %v, want %v", data, tt.data)
			}
		})
	}
}
