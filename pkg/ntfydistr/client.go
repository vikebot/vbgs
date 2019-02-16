package ntfydistr

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/eapache/queue"
	"go.uber.org/zap"
)

// Client represents a single entity responsible for managing notifications to
// a specific user. A Client can hold multiple subscribers (e.g. multiple
// receivers for the user's events.
type Client struct {
	userID   string
	q        *queue.Queue
	qSync    sync.Mutex
	subs     []*subscriber
	subsSync sync.Mutex
}

func newClient(userID string) *Client {
	return &Client{
		userID: userID,
		q:      queue.New(),
		subs:   make([]*subscriber, 0),
	}
}

func (c *Client) addSub(cr *subscriber, log *zap.Logger) {
	// Add subscriber to subs list
	c.subsSync.Lock()
	defer c.subsSync.Unlock()
	c.subs = append(c.subs, cr)
}

func (c *Client) run(stop chan struct{}, log *zap.Logger) {
	log.Debug("starting client notification runner")

	// create ticker for update interval
	tick := time.NewTicker(20 * time.Millisecond)
	for {
		select {
		case <-stop:
			// received stop signal from caller. return
			log.Info("received stop. exiting notification update loop")
			return
		case <-tick.C:
			c.dequeueAndSend(log)
		}
	}
}

func (c *Client) dequeueAndSend(log *zap.Logger) {
	// local buffer for notifications
	var notfs []*event

	// anonymous function to ensure qSync unlock even in panics
	func() {
		c.qSync.Lock()
		defer c.qSync.Unlock()

		// create slice with length of the queue and deque all
		// notifications
		notfs = make([]*event, c.q.Length())
		for i := 0; i < len(notfs); i++ {
			notfs[i] = c.q.Remove().(*event)
		}
	}()

	// no updates for client -> wait for next tick
	if len(notfs) == 0 {
		return
	}

	// marshal notifications slice
	buf, err := json.Marshal(notfs)
	if err != nil {
		log.Error("unable to marshal notification", zap.Error(err))
		return
	}

	// anonymous function to ensure subsSync unlock even in panics
	func() {
		c.subsSync.Lock()
		defer c.subsSync.Unlock()

		disconnSubs := []int{}

		// send client notifications to all subscribers
		for i, s := range c.subs {
			disconnected := s.Send(buf, len(notfs))
			if !disconnected {
				return
			}

			// prepare specific subscriber for closing
			log.Debug("subscriber is disconnecting")
			disconnSubs = append(disconnSubs, i)
			close(s.stop)
		}

		// start at the end of disconnSubs to delete them in a save way
		for i := len(disconnSubs) - 1; i >= 0; i-- {
			// subscriber has disconnected -> remove him from the subscriber
			// list at index i
			c.subs[i] = c.subs[len(c.subs)-1]
			c.subs[len(c.subs)-1] = nil
			c.subs = c.subs[:len(c.subs)-1]
		}
	}()
}

// UserID returns the user id of the user this client represents.
func (c *Client) UserID() string {
	return c.userID
}

// Sub subscribes a new subscriber for all notifications sent to this Client.
// The call is blocking and will only stop once a notification written to the
// SubscriberWriteFunc returns an error and the disconnected return-value
// indicates that this error was caused by a disconnect of the remote party.
// Notifications are queued and send in regular intervals.
func (c *Client) Sub(r Receiver, log *zap.Logger) {
	// Allocate receiver
	cr := newSubscriber(r, make(chan struct{}), log)

	// create dummy client for init of subscriber (so the package caller has
	// the same interface for sending updates during the init as later
	dummy := newClient(c.userID)

	// call init callback with dummy. The subscriber can push everything he
	// wants into the dummy's internal queue-buffer. It won't be dequeued,
	// because noone ever called `run` on the dummy Client.
	r.Init(dummy)

	// in order to send the collected events back to the "real" subscriber
	// we subscribe him to the dummy's event notifications and perform a single
	// dequeueAndSend round
	dummy.addSub(cr, log)
	dummy.dequeueAndSend(log)

	// check if the subscriber already left. if so don't add him to the real
	// Client's live subscriber
	if len(dummy.subs) == 0 {
		return
	}

	// add subscriber to the real client's live subscription list
	c.addSub(cr, log)
	log.Debug("added new subscriber to client")

	// Block till this subscription stops
	<-cr.stop
}

// Push takes any notificationType and data to construct the final
// []byte. Therefore the notification interface must be
// JSON serializable. Push doesn't send anything over the wire. All constructed
// []byte are queued in an internal system and will
// eventually get sent in the next update tick (typically every 20ms). The
// method is safe for concurrent use.
func (c *Client) Push(notificationType string, notification interface{}, log *zap.Logger) {
	if c == nil {
		log.Warn("ntfydistr.Client: client is nil")
		return
	}

	c.qSync.Lock()
	defer c.qSync.Unlock()

	// construct basic packet for sending something into the frontend
	packet := &event{
		Type:  strings.ToLower(notificationType),
		Obj:   notification,
		Unixn: time.Now().UTC().UnixNano(),
	}

	// marshal interface to byte slice
	_, err := json.Marshal(packet)
	if err != nil {
		log.Error("unable to marshal notification", zap.Error(err))
		return
	}

	// queue for pushing to client
	c.q.Add(packet)
}

// PushChat makes it convient to send the message with it's severity level and
// the default prefix to the client. The method is safe for concurrent use.
func (c *Client) PushChat(msg string, sev Severity, log *zap.Logger) {
	c.PushChatPrefixed(defaultChatPrefix, msg, sev, log)
}

// PushChatPrefixed makes it convient to send the message with it's severity
// level and the defined prefix to the client. The method is safe for
// concurrent use.
func (c *Client) PushChatPrefixed(prefix, msg string, sev Severity, log *zap.Logger) {
	c.Push("CHAT", struct {
		Prefix   string   `json:"prefix"`
		Msg      string   `json:"msg"`
		Severity Severity `json:"severity"`
	}{prefix, msg, sev}, log)
}

// PushInitialState makes it convenient to send the initial state (properties)
// to a client. The method is safe for concurrent use.
func (c *Client) PushInitialState(props map[string]string, log *zap.Logger) {
	for k, v := range props {
		c.Push("FLAG", struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{k, v}, log)
	}
}

// PushInfo makes it convenient to send info data (like ip, used sdk, os,
// established indicator, etc.) to a client. The method is safe for concurrent
// use.
func (c *Client) PushInfo(established bool, ip, sdk, sdkLink, os string, log *zap.Logger) {
	type conn struct {
		Established bool   `json:"established"`
		IP          string `json:"ip"`
	}
	type lib struct {
		Name string `json:"name"`
		Link string `json:"link"`
	}
	type obj struct {
		Conn *conn  `json:"conn"`
		Lib  *lib   `json:"lib"`
		OS   string `json:"os"`
	}

	c.Push("INFO", &obj{
		Conn: &conn{established, ip},
		Lib:  &lib{sdk, sdkLink},
		OS:   os,
	}, log)
}
