package ntfydistr

import (
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestLog() *zap.Logger {
	l, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err.Error())
	}

	return l
}

type receiver struct {
	w func([]byte) (bool, error)
}

func (r *receiver) Init(c *Client) {}
func (r *receiver) Write(b []byte) (bool, error) {
	return r.w(b)
}

func funcToReceiver(w func([]byte) (bool, error)) *receiver {
	return &receiver{w}
}

func Test_newSubscriber(t *testing.T) {
	r := funcToReceiver(func(notf []byte) (disconn bool, err error) {
		return true, errors.New("custom test error")
	})
	stop := make(chan struct{})
	log := newTestLog()

	s := newSubscriber(r, stop, log)
	assert.Equal(t, stop, s.stop)
	assert.Equal(t, log, s.log)

	disconn, err := s.r.Write(nil)
	assert.NotNil(t, err)
	assert.Equal(t, "custom test error", err.Error())
	assert.Equal(t, true, disconn)
}

func newTestSub() *subscriber {
	r := funcToReceiver(func(notf []byte) (disconn bool, err error) {
		return true, errors.New("custom test error")
	})
	stop := make(chan struct{})
	log := newTestLog()

	return newSubscriber(r, stop, log)
}

func Test_subscriber_Send(t *testing.T) {
	tests := []struct {
		name             string
		r                Receiver
		wantDisconnected bool
	}{
		{"should be ok", funcToReceiver(func(_ []byte) (bool, error) {
			return false, nil
		}), false},
		{"should be disconnected", funcToReceiver(func(_ []byte) (bool, error) {
			return true, errors.New("pls disconnect")
		}), true},
		{"should be disconnected", funcToReceiver(func(_ []byte) (bool, error) {
			return false, errors.New("pls disconnect")
		}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSub()
			s.r = tt.r

			disc := s.Send([]byte{}, 0)
			assert.Equal(t, tt.wantDisconnected, disc)
		})
	}
}
