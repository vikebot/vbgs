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

func Test_newSubscriber(t *testing.T) {
	swf := func(notf []byte) (disconn bool, err error) {
		return true, errors.New("custom test error")
	}
	stop := make(chan struct{})
	log := newTestLog()

	s := newSubscriber(swf, stop, log)
	assert.Equal(t, stop, s.stop)
	assert.Equal(t, log, s.log)

	disconn, err := s.w([]byte{})
	assert.NotNil(t, err)
	assert.Equal(t, "custom test error", err.Error())
	assert.Equal(t, true, disconn)
}

func newTestSub() *subscriber {
	swf := func(notf []byte) (disconn bool, err error) {
		return true, errors.New("custom test error")
	}
	stop := make(chan struct{})
	log := newTestLog()

	return newSubscriber(swf, stop, log)
}

func Test_subscriber_Send(t *testing.T) {
	tests := []struct {
		name             string
		swf              SubscriberWriteFunc
		wantDisconnected bool
	}{
		{"should be ok", func(_ []byte) (bool, error) {
			return false, nil
		}, false},
		{"should be disconnected", func(_ []byte) (bool, error) {
			return true, errors.New("pls disconnect")
		}, true},
		{"should be disconnected", func(_ []byte) (bool, error) {
			return false, errors.New("pls disconnect")
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSub()
			s.w = tt.swf

			disc := s.Send([]byte{}, 0)
			assert.Equal(t, tt.wantDisconnected, disc)
		})
	}
}
