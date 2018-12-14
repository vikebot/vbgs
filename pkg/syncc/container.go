package syncc

type Container interface {
	NewRequest() Request
}

type Request interface {
	Lock(token string)
	Unlock(token string)
	Aquired(token string) bool
}
