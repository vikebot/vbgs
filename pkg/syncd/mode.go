package syncd

import (
	"bytes"
	"errors"
	"fmt"
)

// Mode is the mode of a Manager instance. For example 'InMem' or 'Etcd'.
type Mode int

const (
	// ModeInMem is the in-memory mode of a Manager. It consists of a in-memory
	// hashmap of token:Mutex pairs.
	ModeInMem Mode = iota
)

// String returns a lower-case ASCII representation of the Mode.
func (m Mode) String() string {
	str, _ := m.string() /* #nosec G104 */
	return str
}

func (m Mode) string() (string, error) {
	switch m {
	case ModeInMem:
		return "inmem", nil
	default:
		return fmt.Sprintf("mode(%d)", m), errors.New("syncd: invalid mode")
	}
}

// MarshalText marshals the Mode to text. Note that the text representation
// drops the 'Mode' prefix (see example).
func (m Mode) MarshalText() ([]byte, error) {
	str, err := m.string()
	if err != nil {
		return nil, err
	}
	return []byte(str), err
}

// UnmarshalText unmarshals text to a Mode. Like MarshalText, UnmarshalText
// expects the text representation of a Mode to drop the 'Mode' prefix (see
// example).
func (m *Mode) UnmarshalText(text []byte) error {
	if m == nil {
		return errors.New("syncd: can't unmarshal a nil *Mode")
	}
	if !m.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("syncd: unrecognized Mode: %q", text)
	}
	return nil
}

func (m *Mode) unmarshalText(text []byte) bool {
	switch string(text) {
	case "inmem", "": // make the zero value useful
		*m = ModeInMem
	default:
		return false
	}
	return true
}
