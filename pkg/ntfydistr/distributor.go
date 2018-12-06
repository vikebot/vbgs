package ntfydistr

import (
	"sync"

	"go.uber.org/zap"
)

// Distributor is the general managing instance for all notifications that
// flow through the system. It stores a list of all clients and is able to
// deliver notifications to them based on different delivery-channels (for
// example: group and broadcast).
type Distributor interface {
	GetClient(userID int) Client
	PushGroup(notificationType string, userIDs []int, notification interface{}, log *zap.Logger)
	PushBroadcast(notificationType string, notification interface{}, log *zap.Logger)
	PushChatGroup(userIDs []int, msg string, sev Severity, log *zap.Logger)
	PushChatPrefixedGroup(userIDs []int, prefix, msg string, sev Severity, log *zap.Logger)
	PushChatBroadcast(msg string, sev Severity, log *zap.Logger)
	PushChatPrefixedBroadcast(prefix, msg string, sev Severity, log *zap.Logger)
}

type dist struct {
	allUserIDs  []int
	clients     map[int]*client
	clientsSync sync.RWMutex
}

// NewDist initializes a new notification Distributor and all it's child
// ClientDistributors.
func NewDist(allUserIDs []int, stop chan struct{}, log *zap.Logger) *dist {
	// create distributor
	d := &dist{
		allUserIDs: allUserIDs,
		clients:    make(map[int]*client, len(allUserIDs)),
	}

	// create all clients
	for _, userID := range allUserIDs {
		// create new Client for userID
		c := newClient(userID)

		// add client to store
		d.clients[userID] = c

		// run client updater
		go c.run(stop, log)
	}

	return d
}

// GetClient returns the Client if currently subscribed. If the client is not
// subscribed nil will be returned. The method is safe for concurrent use.
func (d *dist) GetClient(userID int) Client {
	d.clientsSync.RLock()
	defer d.clientsSync.RUnlock()

	return d.clients[userID]
}

// PushGroup pushes the notification to each member of the group (defined
// through userIDs). The notification interface must be JSON serializable.
// The method is safe for concurrent use.
func (d *dist) PushGroup(notificationType string, userIDs []int, notification interface{}, log *zap.Logger) {
	// lock clients and search all Clients, for which we have userIDs
	d.clientsSync.RLock()
	clients := []Client{}
	for _, userID := range userIDs {
		// lookup client
		if c, ok := d.clients[userID]; ok {
			// add to clients
			clients = append(clients, c)
		}
	}
	d.clientsSync.RUnlock()

	// push notification to each client
	for _, c := range clients {
		c.Push(notificationType, notification, log)
	}
}

// PushBroadcast pushes the notification to all clients. The notification must
// be JSON serializable. The method is safe for concurrent use.
func (d *dist) PushBroadcast(notificationType string, notification interface{}, log *zap.Logger) {
	d.PushGroup(notificationType, d.allUserIDs, notification, log)
}

// PushChatGroup pushes the message with it's severity level and the default
// prefix to all clients listed in userIDs. The method is safe for concurrent
// use.
func (d *dist) PushChatGroup(userIDs []int, msg string, sev Severity, log *zap.Logger) {
	d.PushChatPrefixedGroup(userIDs, "SERVER", msg, sev, log)
}

// PushChatPrefixedGroup pushes the message with it's severity level and the
// defined prefix to all clients listed in userIDs. The method is safe for
// concurrent use.
func (d *dist) PushChatPrefixedGroup(userIDs []int, prefix, msg string, sev Severity, log *zap.Logger) {
	d.PushGroup("CHAT", userIDs, struct {
		Prefix   string   `json:"prefix"`
		Msg      string   `json:"msg"`
		Severity Severity `json:"severity"`
	}{prefix, msg, sev}, log)
}

// PushChatBroadcast pushes the message with it's severity level and the
// default prefix to all clients. The method is safe for concurrent use.
func (d *dist) PushChatBroadcast(msg string, sev Severity, log *zap.Logger) {
	d.PushChatGroup(d.allUserIDs, msg, sev, log)
}

// PushChatPrefixedBroadcast pushes the message with it's severity level and
// the defined prefix to all clients. The method is safe for concurrent use.
func (d *dist) PushChatPrefixedBroadcast(prefix, msg string, sev Severity, log *zap.Logger) {
	d.PushChatPrefixedGroup(d.allUserIDs, prefix, msg, sev, log)
}
