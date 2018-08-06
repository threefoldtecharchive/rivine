package electrum

import (
	"reflect"
	"testing"
)

// Testing protocol argument encoding and decoding implicitly tets protocol version encoding and decoding as well
func TestProtocolArgument_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	type fields struct {
		protocolMin ProtocolVersion
		protocolMax ProtocolVersion
	}

	tests := []struct {
		name    string
		fields  fields
		args    []byte
		wantErr bool
	}{
		{
			name:    "1",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			args:    []byte(`"1.0.0"`),
			wantErr: false,
		},
		{
			name:    "2",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			args:    []byte(`"1.0"`),
			wantErr: false,
		},
		{
			name:    "3",
			fields:  fields{ProtocolVersion{1, 0, 1}, ProtocolVersion{1, 0, 1}},
			args:    []byte(`"1.0.1"`),
			wantErr: false,
		},
		{
			name:    "4",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 1}},
			args:    []byte(`["1.0.0", "1.0.1"]`),
			wantErr: false,
		},
		{
			name:    "5",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			args:    []byte(`"1"`),
			wantErr: true,
		},
		{
			name:    "6",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			args:    []byte(`["1.0", "1"]`),
			wantErr: true,
		},
		{
			name:    "7",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			args:    []byte(`["1.0", "1.0", "1.0"]`),
			wantErr: true,
		},
		{
			name:    "8",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			args:    []byte(`["1.0", "1.0.a"]`),
			wantErr: true,
		},
		{
			name:    "9",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			args:    []byte(`["1.0", "1.0.0.a"]`),
			wantErr: true,
		},
		{
			name:    "10",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			args:    []byte(`"1.a"`),
			wantErr: true,
		},
		{
			name:    "11",
			fields:  fields{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			args:    []byte(`"a.0"`),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := &ProtocolArgument{
				protocolMin: tt.fields.protocolMin,
				protocolMax: tt.fields.protocolMax,
			}
			if err := pa.UnmarshalJSON(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("ProtocolArgument.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProtocolArgument_MarshalJSON(t *testing.T) {
	type fields struct {
		protocolMin ProtocolVersion
		protocolMax ProtocolVersion
	}
	tests := []struct {
		name    string
		input   ProtocolArgument
		want    []byte
		wantErr bool
	}{
		{
			name:    "1",
			input:   ProtocolArgument{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 0}},
			want:    []byte(`"1.0"`),
			wantErr: false,
		},
		{
			name:    "2",
			input:   ProtocolArgument{ProtocolVersion{1, 0, 0}, ProtocolVersion{1, 0, 1}},
			want:    []byte(`["1.0","1.0.1"]`),
			wantErr: false,
		},
		{
			name:    "3",
			input:   ProtocolArgument{ProtocolVersion{1, 1, 0}, ProtocolVersion{1, 1, 1}},
			want:    []byte(`["1.1","1.1.1"]`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := ProtocolArgument{
				protocolMin: tt.input.protocolMin,
				protocolMax: tt.input.protocolMax,
			}
			got, err := pa.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProtocolArgument.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProtocolArgument.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
