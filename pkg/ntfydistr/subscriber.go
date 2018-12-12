package ntfydistr

import "go.uber.org/zap"

// SubscriberWriteFunc is a callback used by the Send operation from a
// Subscriber to send the actual bytes over the wire. The callback should
// return all errors unchanged to the Client. The disconnected return value
// indicates whether or not the returned error is due to a disconnect from
// a remote party. If disconnected is true the current subscription will be
// cancelled and not called again.
type SubscriberWriteFunc func(notf []byte) (disconnected bool, err error)

// Subscriber represents a single entity that wants to receive notifications
// for a specific Client.
type Subscriber interface {
	Send(notfs [][]byte) (disconnected bool)
}

type subscriber struct {
	w    SubscriberWriteFunc
	stop chan struct{}
	log  *zap.Logger
}

func newSubscriber(w SubscriberWriteFunc, stop chan struct{}, log *zap.Logger) *subscriber {
	return &subscriber{
		w:    w,
		stop: stop,
		log:  log,
	}
}

// Send sends all passed []byte notifications and sends
// them to the client using the subscribers SubscriberWriteFunc. Send returns
// whether or not the subscriber has disconnected.
func (s *subscriber) Send(notfs [][]byte) (disconnected bool) {
	// send all notifications to the subscriber
	s.log.Info("sending notifications", zap.Int("amount", len(notfs)))
	for _, nBuf := range notfs {
		// call SubscriberWriteFunc callback
		disconnected, err := s.w(nBuf)
		if err == nil {
			// no error -> continue with next notification
			continue
		}

		// see if the error is a disconnect error
		if disconnected {
			s.log.Info("remote client has forcely closed connection")
			return true
		}

		// error unknown
		s.log.Error("unable to send marshaled notification", zap.Error(err))
	}

	return false
}
