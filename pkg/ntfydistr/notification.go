package ntfydistr

// SerializedNotificationBuffer is a finished notification ready for sending
// over the wire (basically all the information encoded as JSON in a byte
// slice).
type SerializedNotificationBuffer []byte
