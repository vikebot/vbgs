package ntfydistr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		name string
		s    Severity
		want string
	}{
		{"default", SeverityDefault, "default"},
		{"success", SeveritySuccess, "success"},
		{"warning", SeverityWarning, "warning"},
		{"error", SeverityError, "error"},
		{"invalid -2", -2, "severity(-2)"},
		{"invalid -1", -1, "severity(-1)"},
		{"int 0", 0, "default"},
		{"invalid 4", 4, "severity(4)"},
		{"invalid 5000", 5000, "severity(5000)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestSeverity_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       Severity
		want    []byte
		wantErr bool
	}{
		{"default", SeverityDefault, []byte("default"), false},
		{"success", SeveritySuccess, []byte("success"), false},
		{"warning", SeverityWarning, []byte("warning"), false},
		{"error", SeverityError, []byte("error"), false},
		{"int 0", 0, []byte("default"), false},
		{"invalid -2", -2, nil, true},
		{"invalid -1", -1, nil, true},
		{"invalid 4", 4, nil, true},
		{"invalid 5000", 5000, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.MarshalText()
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Equal(t, tt.want, got)
			assert.Nil(t, err)
		})
	}
}

func TestSeverity_MarshalTextNil(t *testing.T) {
	var s Severity
	err := json.Unmarshal([]byte(`{"sev":null}`), &s)
	assert.NotNil(t, err)
}

func TestSeverity_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       Severity
		args    []byte
		wantErr bool
	}{
		{"default", SeverityDefault, []byte("default"), false},
		{"success", SeveritySuccess, []byte("success"), false},
		{"warning", SeverityWarning, []byte("warning"), false},
		{"error", SeverityError, []byte("error"), false},
		{"default", SeverityDefault, []byte(""), false},
		{"invalid", 0, []byte("something other than constants"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Severity
			err := s.UnmarshalText(tt.args)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.s, s)
			}
		})
	}
}

func TestSeverity_UnmarshalTextNil(t *testing.T) {
	var s *Severity
	err := s.UnmarshalText([]byte{})
	assert.NotNil(t, err)
	assert.Equal(t, "can't unmarshal a nil *Severity", err.Error())
}
