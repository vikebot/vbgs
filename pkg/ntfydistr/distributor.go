package ntfydistr

import (
	"sync"

	"go.uber.org/zap"
)

const (
	defaultChatPrefix = "SERVER"
)

// Distributor is the general managing instance for all notifications that
// flow through the system. It stores a list of all clients and is able to
// deliver notifications to them based on different delivery-channels (for
// example: group and broadcast).
type Distributor struct {
	allUserIDs  []string
	clients     map[string]*Client
	clientsSync sync.RWMutex
	stop        chan struct{}
	wg          sync.WaitGroup
}

// NewDistributor initializes a new notification Distributor and all it's child
// ClientDistributors.
func NewDistributor(allUserIDs []string, stop chan struct{}, log *zap.Logger) *Distributor {
	// create distributor
	d := &Distributor{
		allUserIDs: allUserIDs,
		clients:    make(map[string]*Client, len(allUserIDs)),
	}

	// create all clients
	for _, userID := range allUserIDs {
		// create new Client for userID
		c := newClient(userID)

		// add client to store
		d.clients[userID] = c

		// run client updater and increase waitgroup
		d.wg.Add(1)
		go func(userID string) {
			defer d.wg.Done()

			// construct a named child logger
			cLog := log.Named("client" + userID)
			cLog = cLog.With(zap.String("user_id", userID))

			c.run(stop, cLog)
		}(userID)
	}

	return d
}

// Close first waits for the signal of the stop channel provided during
// NewDistributor. Next Close waits for all client update runners to finish.
// As soon as Close returns all started goroutines from ntfydistr should be
// stopped and all remaining messages sent to the subscribers.
func (d *Distributor) Close() {
	<-d.stop
	d.wg.Wait()
}

// GetClient returns the Client if currently subscribed. If the client is not
// subscribed nil will be returned. The method is safe for concurrent use.
func (d *Distributor) GetClient(userID string) *Client {
	d.clientsSync.RLock()
	defer d.clientsSync.RUnlock()

	return d.clients[userID]
}

func (d *Distributor) clientsForUserIDs(userIDs []string) []*Client {
	d.clientsSync.RLock()
	defer d.clientsSync.RUnlock()

	// lock clients and search all, for which we have userIDs
	clients := []*Client{}
	for _, userID := range userIDs {
		// lookup client
		if c, ok := d.clients[userID]; ok {
			// add to clients
			clients = append(clients, c)
		}
	}

	return clients
}

// PushGroup pushes the notification to each member of the group (defined
// through userIDs). The notification interface must be JSON serializable.
// The method is safe for concurrent use.
func (d *Distributor) PushGroup(notificationType string, userIDs []string, notification interface{}, log *zap.Logger) {
	clients := d.clientsForUserIDs(userIDs)

	// push notification to each client
	for _, c := range clients {
		c.Push(notificationType, notification, log)
	}
}

// PushBroadcast pushes the notification to all clients. The notification must
// be JSON serializable. The method is safe for concurrent use.
func (d *Distributor) PushBroadcast(notificationType string, notification interface{}, log *zap.Logger) {
	d.PushGroup(notificationType, d.allUserIDs, notification, log)
}

// PushChatGroup pushes the message with it's severity level and the default
// prefix to all clients listed in userIDs. The method is safe for concurrent
// use.
func (d *Distributor) PushChatGroup(userIDs []string, msg string, sev Severity, log *zap.Logger) {
	d.PushChatPrefixedGroup(userIDs, defaultChatPrefix, msg, sev, log)
}

// PushChatPrefixedGroup pushes the message with it's severity level and the
// defined prefix to all clients listed in userIDs. The method is safe for
// concurrent use.
func (d *Distributor) PushChatPrefixedGroup(userIDs []string, prefix, msg string, sev Severity, log *zap.Logger) {
	clients := d.clientsForUserIDs(userIDs)

	// push notification to each client
	for _, c := range clients {
		c.PushChatPrefixed(prefix, msg, sev, log)
	}
}

// PushChatBroadcast pushes the message with it's severity level and the
// default prefix to all clients. The method is safe for concurrent use.
func (d *Distributor) PushChatBroadcast(msg string, sev Severity, log *zap.Logger) {
	d.PushChatGroup(d.allUserIDs, msg, sev, log)
}

// PushChatPrefixedBroadcast pushes the message with it's severity level and
// the defined prefix to all clients. The method is safe for concurrent use.
func (d *Distributor) PushChatPrefixedBroadcast(prefix, msg string, sev Severity, log *zap.Logger) {
	d.PushChatPrefixedGroup(d.allUserIDs, prefix, msg, sev, log)
}
