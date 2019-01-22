package syncd

import (
	"context"
)

// Request represents a container around Manager that saves it's own in-memory
// acquired-lock-cache, which consists of a token:Mutex hashmap. This cache is
// always checked before calling the parent Manager, which enables features
// like multiple Lock calls for the token, without getting errors.
type Request struct {
	m      Manager
	rLocks map[string]struct{}
	wLocks map[string]struct{}
}

// NewRequest initializes a new Request based on the passed Manager instance.
func NewRequest(m Manager) *Request {
	return &Request{
		m:      m,
		rLocks: make(map[string]struct{}),
		wLocks: make(map[string]struct{}),
	}
}

// Lock locks the write-lock for the token if not already locked. In order to
// acquire the write-lock the parent Manager is called.
//
// Because Request has it's own acquired-lock-cache which is checked before the
// parent Manager is called Lock can be called multiple times without getting
// errors.
func (r *Request) Lock(ctx context.Context, token string) error {
	if _, ok := r.wLocks[token]; ok {
		return nil
	}

	err := r.m.Lock(ctx, token)
	if err != nil {
		return err
	}

	r.wLocks[token] = struct{}{}

	return nil
}

// RLock locks the read-lock for the token if not already locked. In order to
// acquire the read-lock the parent Manager is called.
//
// Because Request has it's own acquired-lock-cache which is checked before the
// parent Manager is called RLock can be called multiple times without getting
// errors.
func (r *Request) RLock(ctx context.Context, token string) error {
	if _, ok := r.rLocks[token]; ok {
		return nil
	}

	err := r.m.RLock(ctx, token)
	if err != nil {
		return err
	}

	r.rLocks[token] = struct{}{}

	return nil
}

// Unlock unlocks the write-lock for the token if not already locked. In order
// to release the write-lock the parent Manager is called.
//
// Because Request has it's own acquired-lock-cache which is checked before the
// parent Manager is called Unlock can be called multiple times without getting
// errors.
func (r *Request) Unlock(ctx context.Context, token string) error {
	if _, ok := r.wLocks[token]; !ok {
		return nil
	}

	err := r.m.Unlock(ctx, token)
	if err != nil {
		return err
	}

	delete(r.wLocks, token)

	return nil
}

// RUnlock unlocks the read-lock for the token if not already locked. In order
// to release the read-lock the parent Manager is called.
//
// Because Request has it's own acquired-lock-cache which is checked before the
// parent Manager is called RUnlock can be called multiple times without
// getting errors.
func (r *Request) RUnlock(ctx context.Context, token string) error {
	if _, ok := r.rLocks[token]; !ok {
		return nil
	}

	err := r.m.RUnlock(ctx, token)
	if err != nil {
		return err
	}

	delete(r.rLocks, token)

	return nil
}

// Acquired indicates if the the write-lock for the token was already acquired.
// The check is performed using the acquired-lock-cache.
func (r *Request) Acquired(token string) bool {
	if _, ok := r.wLocks[token]; ok {
		return true
	}

	return false
}

// RAcquired indicates if the the read-lock for the token was already acquired.
// The check is performed using the acquired-lock-cache.
func (r *Request) RAcquired(token string) bool {
	if _, ok := r.rLocks[token]; ok {
		return true
	}

	return false
}

// UnlockAll unlocks all read- and write-locks. The Unlocks itself are
// performed with the parent Manager. Which tokens were acquired is determined
// via the acquired-lock-cache. All errors returned from the parent Manager
// are sent to the unbuffered error channel (Attention: this means you need
// to receive on the channel. If you don't receive errors sent to it, the
// application will block.). After one Unlock attempt foreach previously
// acquired token is made, the channel is closed.
//
// Example usage:
//     for err := r.UnlockAll(context.Background()) {
//         fmt.Println(err)
//     }
func (r *Request) UnlockAll(ctx context.Context) chan error {
	errChan := make(chan error)

	go func() {
		defer close(errChan)

		for token := range r.rLocks {
			err := r.RUnlock(ctx, token)
			if err != nil {
				errChan <- err
			}
		}

		for token := range r.wLocks {
			err := r.Unlock(ctx, token)
			if err != nil {
				errChan <- err
			}
		}
	}()

	return errChan
}
