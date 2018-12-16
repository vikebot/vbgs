package syncd

import (
	"context"
)

// Request represents a container around Manager that saves it's own in-memory
// aquired lock-token chache. This cache is always checked before calling
// the parent Manager, which enables features like multiple Lock calls for the
// token, without getting errors.
type Request struct {
	Mutexd
	m    Manager
	aquR map[string]struct{}
	aquW map[string]struct{}
}

func newRequest(m Manager) *Request {
	return &Request{
		m:    m,
		aquR: make(map[string]struct{}),
		aquW: make(map[string]struct{}),
	}
}

// Lock locks the write-lock for the token if not already locked. In order to
// aquire the write-lock the parent Manager is called.
//
// Because Request has it's own lock-aquisition cache which is checked before
// the parent Manager is called Lock can be called multiple times without
// getting errors.
func (r *Request) Lock(ctx context.Context, token string) error {
	if _, ok := r.aquW[token]; ok {
		return nil
	}

	err := r.m.Lock(ctx, token)
	if err != nil {
		return err
	}

	r.aquW[token] = struct{}{}

	return nil
}

// RLock locks the read-lock for the token if not already locked. In order to
// aquire the read-lock the parent Manager is called.
//
// Because Request has it's own lock-aquisition cache which is checked before
// the parent Manager is called RLock can be called multiple times without
// getting errors.
func (r *Request) RLock(ctx context.Context, token string) error {
	if _, ok := r.aquR[token]; ok {
		return nil
	}

	err := r.m.RLock(ctx, token)
	if err != nil {
		return err
	}

	r.aquR[token] = struct{}{}

	return nil
}

// Unlock unlocks the write-lock for the token if not already locked. In order
// to release the write-lock the parent Manager is called.
//
// Because Request has it's own lock-aquisition cache which is checked before
// the parent Manager is called Unlock can be called multiple times without
// getting errors.
func (r *Request) Unlock(ctx context.Context, token string) error {
	if _, ok := r.aquW[token]; !ok {
		return nil
	}

	err := r.m.Unlock(ctx, token)
	if err != nil {
		return err
	}

	delete(r.aquW, token)

	return nil
}

// RUnlock unlocks the read-lock for the token if not already locked. In order
// to release the read-lock the parent Manager is called.
//
// Because Request has it's own lock-aquisition cache which is checked before
// the parent Manager is called RUnlock can be called multiple times without
// getting errors.
func (r *Request) RUnlock(ctx context.Context, token string) error {
	if _, ok := r.aquR[token]; !ok {
		return nil
	}

	err := r.m.RUnlock(ctx, token)
	if err != nil {
		return err
	}

	delete(r.aquR, token)

	return nil
}

// Aquired indicates if the the write-lock for the token was already aquired.
// The check is performed using the lock-aquistition cache.
func (r *Request) Aquired(token string) bool {
	if _, ok := r.aquW[token]; ok {
		return true
	}

	return false
}

// RAquired indicates if the the read-lock for the token was already aquired.
// The check is performed using the lock-aquistition cache.
func (r *Request) RAquired(token string) bool {
	if _, ok := r.aquR[token]; ok {
		return true
	}

	return false
}

// UnlockAll unlocks all read- and write-locks. The Unlocks itself are
// performed with the parent Manager. Which token's we aquired is determined
// via the lock-aquisition cache. All errors returned from the parent Mangager
// are sent to the unbuffered error channel (Attention: this means you need
// to receive on the channel. If you don't receive errors the application will
// block.). After a Unlock attempt foreach previously aquired token is made the
// channel is closed.
//
// Example usage:
//     for err := r.UnlockAll(context.Background()) {
//         fmt.Println(err)
//     }
func (r *Request) UnlockAll(ctx context.Context) chan error {
	errChan := make(chan error)

	go func() {
		for token := range r.aquR {
			err := r.RUnlock(ctx, token)
			if err != nil {
				errChan <- err
			}
		}

		for token := range r.aquW {
			err := r.Unlock(ctx, token)
			if err != nil {
				errChan <- err
			}
		}

		close(errChan)
	}()

	return errChan
}
