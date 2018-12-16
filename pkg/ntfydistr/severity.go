package ntfydistr

import (
	"bytes"
	"errors"
	"fmt"
)

// Severity is the severity level used for a specific chat message. Depending
// on the client's implementation this resulsts in a prefix or special colors.
type Severity int

const (
	// SeverityDefault should be used for all chat messages that don't fit into
	// one of the other Severity levels.
	SeverityDefault Severity = iota
	// SeveritySuccess should be used to indicate the success of something.
	SeveritySuccess
	// SeverityWarning should be used to indicate that something didn't went as
	// expected, but the failure could be handled gracefully.
	SeverityWarning
	// SeverityError should be used to indicate the error of something.
	SeverityError
)

// String returns a lower-case ASCII representation of the severity.
func (s Severity) String() string {
	str, _ := s.string() /* #nosec G104 */
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
	if err != nil {
		return nil, err
	}
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
