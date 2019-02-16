package ntfydistr

import "go.uber.org/zap"

type subscriber struct {
	r    Receiver
	stop chan struct{}
	log  *zap.Logger
}

func newSubscriber(r Receiver, stop chan struct{}, log *zap.Logger) *subscriber {
	return &subscriber{
		r:    r,
		stop: stop,
		log:  log,
	}
}

// Send sends the passed []byte notification and sends it to the client using
// the subscribers SubscriberWriteFunc. Send returns whether or not the
// subscriber has disconnected.
func (s *subscriber) Send(buffer []byte, amount int) (disconnected bool) {
	// send all notifications to the subscriber
	s.log.Debug("sending notifications", zap.Int("amount", amount), zap.Int("buffer_len", len(buffer)))

	// call SubscriberWriteFunc callback
	disconnected, err := s.r.Write(buffer)
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
