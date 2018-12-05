package ntfydistr

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/eapache/queue"
	"github.com/vikebot/vbcore"
	"go.uber.org/zap"
)

// SerializedNotificationBuffer is a finished notification ready for sending
// over the wire (basically all the information encoded as JSON in a byte
// slice).
type SerializedNotificationBuffer []byte

// ClientWriteFunc is a callback used by the running operation from a
// ClientDistributor to send the actual bytes over the wire. The callback
// should return all errors unchanged to the ClientDistributor, because
// the error types are checked (therefore important for the correct control
// flow).
type ClientWriteFunc func(notf SerializedNotificationBuffer) error

// ClientIsDisconnectedErrorFunc is a callback used by the running operation
// from the ClientDistributor to determine whether or not an error returned
// from a ClientWriteFunc due to a disconnect from the remote party.
type ClientIsDisconnectedErrorFunc func(err error) bool

// ClientDistributor represents a single entity responsible for managing
// notifications to a specific client.
type ClientDistributor interface {
	UserID() int
	Run(stop chan struct{}, w ClientWriteFunc, isDisconnectedErr ClientIsDisconnectedErrorFunc, log *zap.Logger)
	Push(notificationType string, notification interface{}, log *zap.Logger)
	PushInitialState(props map[string]string, user *vbcore.SafeUser, log *zap.Logger)
	PushInfo(established bool, ip, sdk, sdkLink, os string, log *zap.Logger)
}

type cdist struct {
	userID   int
	syncRoot sync.Mutex
	running  bool
	q        *queue.Queue
}

func newCdist(userID int) *cdist {
	return &cdist{
		userID: userID,
		q:      queue.New(),
	}
}

func (c *cdist) UserID() int {
	return c.userID
}

// Run starts the client notification update runner. It dequeues (in a regular
// interval) all internally queued messages and sends it with the
// ClientWriteFunc callback to the client. The call is blocking and can only
// be stopped with either the client disconnecting or sending to (or closing)
// the stop channel.
func (c *cdist) Run(stop chan struct{}, w ClientWriteFunc, isDisconnectedErr ClientIsDisconnectedErrorFunc, log *zap.Logger) {
	// set client to running to active
	c.running = true

	// set deferred function deactivating the running state
	defer func() {
		c.running = false
	}()

	// create ticker for update interval
	tick := time.NewTicker(20 * time.Millisecond)
	for {
		select {
		case <-stop:
			// received stop signal from caller. return
			log.Info("received stop. exiting notification update loop")
			return
		case <-tick.C:
			// local buffer for notifications
			var notfs []SerializedNotificationBuffer

			// anonymous function to ensure syncRoot unlock even in panics
			func() {
				c.syncRoot.Lock()
				defer c.syncRoot.Unlock()

				// create slice with length of the queue and deque all
				// notifications
				notfs = make([]SerializedNotificationBuffer, c.q.Length())
				for i := 0; i < len(notfs); i++ {
					notfs[i] = c.q.Remove().(SerializedNotificationBuffer)
				}
			}()

			// no updates for client -> wait for next tick
			if len(notfs) == 0 {
				continue
			}

			// send all dequeued notifications to client
			log.Info("sending notifications", zap.Int("amount", len(notfs)))
			for _, nBuf := range notfs {
				err := w(nBuf)
				if err == nil {
					// no error -> continue with next notification
					continue
				}

				// see if the error is a disconnect error
				if isDisconnectedErr(err) {
					log.Info("remote client has forcely closed connection")
					return
				}

				// error unknown
				log.Error("unable to send marshaled notification", zap.Error(err))
			}
		}
	}
}

// Push takes any notificationType and data to construct the final
// SerializedNotificationBuffer. Therefore the notification interface must be
// JSON serializable.It doesn't send anything over the wire. All constructed
// SerializedNotificationBuffer are queued in an internal system and will
// eventually get sent in the next update tick (typically every 20ms). The
// method is safe for concurrent use.
func (c *cdist) Push(notificationType string, notification interface{}, log *zap.Logger) {
	c.syncRoot.Lock()
	defer c.syncRoot.Unlock()

	// return if the client is not connected at all
	if !c.running {
		return
	}

	// construct basic packet for sending something into the frontend
	packet := struct {
		Type  string      `json:"type"`
		Obj   interface{} `json:"obj"`
		Unixn int64       `json:"unixn"`
	}{
		Type:  notificationType,
		Obj:   notification,
		Unixn: time.Now().UTC().UnixNano(),
	}

	// marshal interface to byte slice
	buf, err := json.Marshal(packet)
	if err != nil {
		log.Error("unable to marshal notification", zap.Error(err))
		return
	}

	// queue for pushing to client
	c.q.Add(buf)
}

// PushInitialState makes it convenient to send the initial state (properties
// and user data) to a client. The method is safe for concurrent use.
func (c *cdist) PushInitialState(props map[string]string, user *vbcore.SafeUser, log *zap.Logger) {
	c.pushInitialStateProps(props, log)
	c.pushInitialStateUser(user, log)
}

func (c *cdist) pushInitialStateProps(props map[string]string, log *zap.Logger) {
	for k, v := range props {
		c.Push("FLAG", struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{k, v}, log)
	}
}

func (c *cdist) pushInitialStateUser(user *vbcore.SafeUser, log *zap.Logger) {
	type objUser struct {
		Name       string `json:"name"`
		Username   string `json:"username"`
		Picture    string `json:"picture"`
		Permission string `json:"permission"`
	}
	type obj struct {
		User *objUser `json:"user"`
	}

	c.Push("USERINFO", &obj{
		User: &objUser{user.Name, user.Username, "", user.PermissionString},
	}, log)
}

// PushInfo makes it convenient to send info data (like ip, used sdk, os,
// established indicator, etc.) to a client. The method is safe for concurrent
// use.
func (c *cdist) PushInfo(established bool, ip, sdk, sdkLink, os string, log *zap.Logger) {
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
		Lib: &lib{sdk, sdkLink},
		OS: os,
	}, log)
}

// Distributor is the general managing instance for all notifications that
// flow through the system. It stores a list of all subscribed clients and
// is able to deliver notifications to them based on different delivery-
// channels (for example: group and broadcast).
type Distributor interface {
	GetClient(userID int) ClientDistributor
	PushGroup(notificationType string, userIDs []int, notification interface{}, log *zap.Logger)
	PushBroadcast(notificationType string, notification interface{}, log *zap.Logger)
	PushChatGroup(userIDs []int, msg string, sev Severity, log *zap.Logger)
	PushChatPrefixedGroup(userIDs []int, prefix, msg string, sev Severity, log *zap.Logger)
	PushChatBroadcast(msg string, sev Severity, log *zap.Logger)
	PushChatPrefixedBroadcast(prefix, msg string, sev Severity, log *zap.Logger)
}

type dist struct {
	allUserIDs   []int
	subs         map[int]ClientDistributor
	subsSyncRoot sync.RWMutex
}

// NewDist initializes a new notification Distributor and all it's child
// ClientDistributors.
func NewDist(allUserIDs []int) *dist {
	// create distributor
	d := &dist{
		allUserIDs: allUserIDs,
		subs:       make(map[int]ClientDistributor, len(allUserIDs)),
	}

	// create all client distributors
	for _, userID := range allUserIDs {
		d.subs[userID] = newCdist(userID)
	}

	return d
}

// GetClient returns the ClientDistributor if currently subscribed. If the
// client is not subscribed nil will be returned. The method is safe for
// concurrent use.
func (d *dist) GetClient(userID int) ClientDistributor {
	d.subsSyncRoot.RLock()
	defer d.subsSyncRoot.RUnlock()

	return d.subs[userID]
}

// PushGroup pushes the notification to each member of the group (defined
// through userIDs) if they are connected. The notification interface must
// be JSON serializable. The method is safe for concurrent use.
func (d *dist) PushGroup(notificationType string, userIDs []int, notification interface{}, log *zap.Logger) {
	// lock subscribed group and lookup all userIDs to their associated
	// ClientDistributor
	d.subsSyncRoot.RLock()
	cds := make([]ClientDistributor, len(userIDs))
	for id := range userIDs {
		if idD, ok := d.subs[id]; ok {
			cds = append(cds, idD)
		}
	}
	d.subsSyncRoot.RUnlock()

	// push notification to each client
	for _, cd := range cds {
		cd.Push(notificationType, notification, log)
	}
}

// PushBroadcast pushes the notification to all subscribed clients. The
// notification must be JSON serializable. The method is safe for concurrent
// use.
func (d *dist) PushBroadcast(notificationType string, notification interface{}, log *zap.Logger) {
	d.PushGroup(notificationType, d.allUserIDs, notification, log)
}

// PushChatGroup pushes the message with it's severity level and the default
// prefix to all connected clients listed in userIDs. The method is safe for
// concurrent use.
func (d *dist) PushChatGroup(userIDs []int, msg string, sev Severity, log *zap.Logger) {
	d.PushChatPrefixedGroup(userIDs, "SERVER", msg, sev, log)
}

// PushChatPrefixedGroup pushes the message with it's severity level and the
// defined prefix to all connected clients listed in userIDs. The method is
// safe for concurrent use.
func (d *dist) PushChatPrefixedGroup(userIDs []int, prefix, msg string, sev Severity, log *zap.Logger) {
	d.PushGroup("CHAT", userIDs, struct {
		Prefix   string   `json:"prefix"`
		Msg      string   `json:"msg"`
		Severity Severity `json:"severity"`
	}{prefix, msg, sev}, log)
}

// PushChatBroadcast pushes the message with it's severity level and the
// default prefix to all connected clients. The method is safe for concurrent
// use.
func (d *dist) PushChatBroadcast(msg string, sev Severity, log *zap.Logger) {
	d.PushChatGroup(d.allUserIDs, msg, sev, log)
}

// PushChatPrefixedBroadcast pushes the message with it's severity level and
// the defined prefix to all connected clients. The method is safe for
// concurrent use.
func (d *dist) PushChatPrefixedBroadcast(prefix, msg string, sev Severity, log *zap.Logger) {
	d.PushChatPrefixedGroup(d.allUserIDs, prefix, msg, sev, log)
}
