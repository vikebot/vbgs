package ntfydistr

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
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
		{"invalid 4", 4,  nil, true},
		{"invalid 5000", 5000, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.MarshalText()
			if !tt.wantErr {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSeverity_MarshalTextNil(t *testing.T) {
	var s Severity
	err := json.Unmarshal([]byte(`{"sev":null}`), &s)
	assert.NotNil(t, err)
}

func TestSeverity_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		s       *Severity
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.UnmarshalText(tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("Severity.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSeverity_unmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name string
		s    *Severity
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.unmarshalText(tt.args.text); got != tt.want {
				t.Errorf("Severity.unmarshalText() = %v, want %v", got, tt.want)
			}
		})
	}
}
