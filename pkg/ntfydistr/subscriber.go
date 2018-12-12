package ntfydistr

import "go.uber.org/zap"

// SubscriberWriteFunc is a callback used by the Send operation from a
// Subscriber to send the actual bytes over the wire. The callback should
// return all errors unchanged to the Client. The disconnected return value
// indicates whether or not the returned error is due to a disconnect from
// a remote party. If disconnected is true the current subscription will be
// cancelled and not called again.
type SubscriberWriteFunc func(notf []byte) (disconnected bool, err error)

// SubscriberInitFunc is a callbacked used during initializing a new
// Subscriber. It is called with a dummy client which collects all notifcations
// without sending them in the background. The intention of this method is to
// send informations only needed by the newly connected subscriber.
type SubscriberInitFunc func(c Client)

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

// Send sends the passed []byte notification and sends it to the client using
// the subscribers SubscriberWriteFunc. Send returns whether or not the
// subscriber has disconnected.
func (s *subscriber) Send(buffer []byte, amount int) (disconnected bool) {
	// send all notifications to the subscriber
	s.log.Info("sending notifications", zap.Int("amount", amount), zap.Int("buffer_len", len(buffer)))

	// call SubscriberWriteFunc callback
	disconnected, err := s.w(buffer)
	if err == nil {
		// no error -> return
		return false
	}

	// see if the error is a disconnect error
	if disconnected {
		s.log.Info("remote client has forcely closed connection")
		return true
	}

	// error unknown
	s.log.Error("unable to send marshaled notification", zap.Error(err))
	return false
}
