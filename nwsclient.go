package main

import (
	"sync"

	"github.com/eapache/queue"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type nwsclient struct {
	WSRqID   string
	UserID   int
	Mt       int
	Ws       *websocket.Conn
	Queue    *queue.Queue
	SyncRoot sync.Mutex
	Log      *zap.Logger
}

func (c *nwsclient) Write(buf []byte) error {
	return c.Ws.WriteMessage(c.Mt, buf)
}

func (c *nwsclient) WriteStr(str string) error {
	return c.Write([]byte(str))
}

func (c *nwsclient) Notify(u update) {
	c.SyncRoot.Lock()
	defer c.SyncRoot.Unlock()

	c.Queue.Add(u)
}
