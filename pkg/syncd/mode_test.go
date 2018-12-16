package syncd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMode_String(t *testing.T) {
	tests := []struct {
		name string
		m    Mode
		want string
	}{
		{"inmem", ModeInMem, "inmem"},
		{"int 0", 0, "inmem"},
		{"invalid", -1, "mode(-1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.m.String())
		})
	}
}

func TestMode_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		m       Mode
		want    []byte
		wantErr bool
	}{
		{"inmem", ModeInMem, []byte("inmem"), false},
		{"int 0", 0, []byte("inmem"), false},
		{"invalid", -1, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalText()
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Equal(t, tt.want, got)
			assert.Nil(t, err)
		})
	}
}

func TestMode_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		m       Mode
		args    []byte
		wantErr bool
	}{
		{"inmem", ModeInMem, []byte("inmem"), false},
		{"inmem", ModeInMem, []byte(""), false},
		{"invalid", 0, []byte("invalid mode representation"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Mode
			err := m.UnmarshalText(tt.args)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.m, m)
			}
		})
	}
}

func TestMode_UnmarshalTextNil(t *testing.T) {
	var m *Mode
	err := m.UnmarshalText([]byte("inmem"))
	assert.NotNil(t, err)
	assert.Equal(t, "syncd: can't unmarshal a nil *Mode", err.Error())
}

type jsonTest struct {
	Mode Mode `json:"mode"`
}

func TestMode_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		want    Mode
		wantErr bool
	}{
		{"inmem", `{"mode":"inmem"}`, ModeInMem, false},
		{"inmem by default", `{"mode":""}`, ModeInMem, false},
		{"inmem", `{"mode":null}`, ModeInMem, false},
		{"invalid", `{"mode":"invalid mode representation"}`, ModeInMem, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj jsonTest
			err := json.Unmarshal([]byte(tt.jsonStr), &obj)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, obj.Mode)
		})
	}
}

func TestMode_JSONMarshal(t *testing.T) {
	tests := []struct {
		name    string
		m       Mode
		want    string
		wantErr bool
	}{
		{"inmem", ModeInMem, `{"mode":"inmem"}`, false},
		{"invalid", -1, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := jsonTest{
				Mode: tt.m,
			}

			got, err := json.Marshal(obj)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}
