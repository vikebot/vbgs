package syncd

import "context"

// Mutexd defines a simple interface for a network-ready Mutex implementation.
// In contrast to to sync.Mutex structs Mutexd accepts a context.Context as
// first argument, the token which should be acquired and returns an error that
// may be caused due to network calls, etc.
type Mutexd interface {
	Lock(ctx context.Context, token string) error
	RLock(ctx context.Context, token string) error
	Unlock(ctx context.Context, token string) error
	RUnlock(ctx context.Context, token string) error
}
