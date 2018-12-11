package ntfydistr

import (
	"bytes"
	"errors"
	"fmt"
)

type Severity int

const (
	SeverityDefault Severity = iota
	SeveritySuccess
	SeverityWarning
	SeverityError
)

// String returns a lower-case ASCII representation of the severity.
func (s Severity) String() string {
	str, _ := s.string()
	return str
}

func (s Severity) string() (string, error) {
	switch s {
	case SeverityDefault:
		return "default", nil
	case SeveritySuccess:
		return "success", nil
	case SeverityWarning:
		return "warning", nil
	case SeverityError:
		return "error", nil
	default:
		return fmt.Sprintf("severity(%d)", s), errors.New("ntfydistr: invalid severity")
	}
}

// MarshalText marshals the Severity to text. Note that the text representation
// drops the -Severity prefix (see example).
func (s Severity) MarshalText() ([]byte, error) {
	str, err := s.string()
	return []byte(str), err
}

// UnmarshalText unmarshals text to a Severity. Like MarshalText, UnmarshalText
// expects the text representation of a Level to drop the -Severity prefix (see
// example).
func (s *Severity) UnmarshalText(text []byte) error {
	if s == nil {
		return errors.New("can't unmarshal a nil *Severity")
	}
	if !s.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized severity: %q", text)
	}
	return nil
}

func (s *Severity) unmarshalText(text []byte) bool {
	switch string(text) {
	case "default", "": // make the zero value useful
		*s = SeverityDefault
	case "success":
		*s = SeveritySuccess
	case "warning":
		*s = SeverityWarning
	case "error":
		*s = SeverityError
	default:
		return false
	}
	return true
}
