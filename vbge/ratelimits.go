package vbge

var (
	// Throttles are integers that describe how often a specific command
	// can be called PER SECOND from a client. Therefore a 1 means the client
	// is able to call this operation each second once. A 2 means the client
	// could use the operation twice a second. And so on.
	rotateThrottle      = 2
	moveThrottle        = 1
	attackThrottle      = 3
	radarThrottle       = 1
	watchThrottle       = 2
	environmentThrottle = 4
	scoutThrottle       = 2
	defendThrottle      = 1
	healthThrottle      = 2
)
