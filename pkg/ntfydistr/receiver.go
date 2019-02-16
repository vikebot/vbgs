package ntfydistr

// Receiver represents a single entity that wants to subscribe to notifications
// for a specific Client.
type Receiver interface {
	// Init is a used during initializing a new Receiver. It is called with a
	// dummy client which collects all notifcations without sending them in the
	// background. The intention of this method is to send informations only
	// needed by the newly connected Receiver, but not all the other's already
	// subscribed to the user.
	Init(c *Client)

	// Write is used by the Send operation from a subscriber to send the actual
	// bytes over the wire. The function should return all errors unchanged to
	// the Client. The disconnected return value indicates whether or not the
	// returned error is due to a disconnect from a remote party. If
	// disconnected is true the current subscription will be cancelled and not
	// called again.
	Write(notf []byte) (disconnected bool, err error)
}
