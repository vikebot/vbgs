package main

import (
	"fmt"
	"sync"
)

// ---------------------------------------------------------------------------

type regntcp struct {
	m     map[int]*ntcpclient
	baton sync.Mutex
}

func (r *regntcp) Put(c *ntcpclient) error {
	r.baton.Lock()
	defer r.baton.Unlock()

	if _, ok := r.m[c.UserID]; ok {
		return fmt.Errorf("user(%d) already exists in registry 'regntcp'", c.UserID)
	}
	r.m[c.UserID] = c

	return nil
}

func (r *regntcp) Get(userID int) *ntcpclient {
	r.baton.Lock()
	defer r.baton.Unlock()

	return r.m[userID]
}

func (r *regntcp) Delete(c *ntcpclient) {
	r.baton.Lock()
	defer r.baton.Unlock()

	delete(r.m, c.UserID)
}

// ---------------------------------------------------------------------------

type regnws struct {
	m     map[int][]*nwsclient
	baton sync.Mutex
}

func (r *regnws) Put(c *nwsclient) {
	r.baton.Lock()
	defer r.baton.Unlock()

	r.m[c.UserID] = append(r.m[c.UserID], c)
}

func (r *regnws) Get(userID int) []*nwsclient {
	r.baton.Lock()
	defer r.baton.Unlock()

	return r.m[userID]
}

func (r *regnws) Delete(c *nwsclient) {
	r.baton.Lock()
	defer r.baton.Unlock()

	if len(r.m[c.UserID]) == 1 {
		delete(r.m, c.UserID)
		return
	}

	for i, ws := range r.m[c.UserID] {
		if ws != c {
			continue
		}

		// get slices of user clients
		a := r.m[c.UserID]

		// remove connection at index i
		a[i] = a[len(a)-1]
		a[len(a)-1] = nil
		a = a[:len(a)-1]

		// set connections
		r.m[c.UserID] = a

		return
	}
}

// ---------------------------------------------------------------------------

var ntcpRegistry regntcp
var nwsRegistry regnws

func registryInit() {
	ntcpRegistry = regntcp{
		m: map[int]*ntcpclient{},
	}
	nwsRegistry = regnws{
		m: map[int][]*nwsclient{},
	}
	logctx.Info("initialized registry storages for ntcpclient and nwsclient structs")
}
