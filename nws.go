package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/eapache/queue"
	"github.com/gorilla/websocket"
	"github.com/vikebot/vbcore"
	"github.com/vikebot/vbdb"
	"github.com/vikebot/vbgs/vbge"
	"go.uber.org/zap"
)

var nwsUpgrader websocket.Upgrader

func nwsInit(start chan bool, shutdown chan bool) {
	nwsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header["Origin"]
			if len(origin) == 0 {
				return false
			}
			u, err := url.Parse(origin[0])
			if err != nil {
				return false
			}
			return u.Host == config.Network.WS.ValidOrigin
		},
	}

	srv := &http.Server{Addr: config.Network.WS.Addr}
	http.HandleFunc("/", nwsHandler)

	go func() {
		// Wait for start signal
		logctx.Info("nws ready. waiting for start signal")
		<-start

		go nwsRun(srv)

		// Shutdown websocket when signal is received
		<-shutdown
		err := srv.Shutdown(nil)
		if err != nil {
			logctx.Warn("nws shutdown failed", zap.Error(err))
		}
	}()
}

func nwsRun(srv *http.Server) {
	var srvErr error

	logctx.Info("accepting clients on nws listener")
	if config.Network.WS.TLS.Active {
		srvErr = srv.ListenAndServeTLS(config.Network.WS.TLS.Cert, config.Network.WS.TLS.PKey)
	} else {
		srvErr = srv.ListenAndServe()
	}

	if srvErr != nil {
		logctx.Fatal("nws listen failed", zap.Error(srvErr))
	}
}

func nwsHandler(w http.ResponseWriter, r *http.Request) {
	wsrqid := vbcore.FastRandomString(32)
	c := &nwsclient{
		WSRqID: wsrqid,
		Ctx:    logctx.With(zap.String("wsrqid", wsrqid)),
	}

	c.Ctx.Info("connected", zap.String("ip", r.RemoteAddr))

	ws, err := nwsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		c.Ctx.Error("failed to upgrade http connection", zap.Error(err))
		return
	}
	defer func() {
		c.Ctx.Info("closed", zap.String("ip", r.RemoteAddr))
		err = ws.Close()
		if err != nil {
			c.Ctx.Error("error during closing websocket", zap.Error(err))
		}
	}()

	c.Ws = ws

	nws(c)
}

func nws(c *nwsclient) {
	mt, watchtoken, err := c.Ws.ReadMessage()
	if err != nil {
		c.Ctx.Warn("failed reading message from websocket", zap.Error(err))
		return
	}
	c.Mt = mt

	v, exists, success := vbdb.RoundentryFromWatchtokenCtx(string(watchtoken), c.Ctx)
	if !success {
		err = c.WriteStr("Internal server error")
		if err != nil {
			c.Ctx.Warn("unable to send internal server error message", zap.Error(err))
		}
		return
	}
	if !exists {
		c.Ctx.Warn("client provided unknown watchtoken", zap.String("watchtoken", string(watchtoken)))
		err = c.WriteStr("Unknown watchtoken")
		if err != nil {
			c.Ctx.Warn("unable to send unknown watchtoken error message", zap.Error(err))
		}
		return
	}
	c.UserID = v.UserID
	c.Ctx.Info("authenticated", zap.Int("user_id", v.UserID))

	if config.Battle.RoundID != v.RoundID {
		c.Ctx.Warn("valid watchtoken references invalid round",
			zap.Int("config_round_id", config.Battle.RoundID),
			zap.Int("watchtoken_round_id", v.RoundID))

		err = c.WriteStr("Internal server error")
		if err != nil {
			c.Ctx.Warn("unable to send internal server error message", zap.Error(err))
		}
		return
	}

	c.Queue = queue.New()
	nwsRegistry.Put(c)

	defer func() {
		nwsRegistry.Delete(c)
		c.Ctx.Debug("disconnected. removed from registry")
	}()

	if config.Network.WS.Flags.Debug {
		updateDist.PushTypeFlag(c, "debug", true)
		c.Ctx.Debug("sending debug flag to nwsclient")
	}

	// send user info
	updateDist.PushTypeUserinfo(c)

	// initialGame is a struct for the first message
	// in an ws connection to init the game in
	// vbwatch
	type initialGame struct {
		TotalMapsize    vbge.Location        `json:"totalmapsize"`
		ViewableMapsize vbge.Location        `json:"viewablemapsize"`
		MaxHealth       int                  `json:"maxhealth"`
		PlayerMapentity [][]*vbge.EntityResp `json:"playermapentity"`
		Startplayer     string               `json:"startplayer"`
	}

	var player = battle.Players[c.UserID]

	viewableMapsize := vbge.Location{
		X: vbge.RenderWidth,
		Y: vbge.RenderHeight,
	}

	playerMapentity, err := vbge.GetViewableMapentity(viewableMapsize.X, viewableMapsize.Y, c.UserID, battle, true)
	if err != nil {
		c.Ctx.Error("failed getting mapentity", zap.Error(err))
		return
	}

	init := &initialGame{
		TotalMapsize: vbge.Location{
			X: vbge.MapWidth,
			Y: vbge.MapHeight,
		},
		ViewableMapsize: viewableMapsize,
		MaxHealth:       vbge.MaxHealth,
		Startplayer:     player.GRenderID,
		PlayerMapentity: playerMapentity.Matrix,
	}

	initObj, err := json.Marshal(init)
	if err != nil {
		c.Ctx.Error("failed sending message to websocket connection", zap.Error(err))
		return
	}

	updateDist.PushInit(c, initObj)
	c.Ctx.Debug("sending init package to nwsclient")

	for {
		time.Sleep(time.Millisecond * 100)

		var updates []update
		func() {
			c.SyncRoot.Lock()
			defer c.SyncRoot.Unlock()

			updates = make([]update, c.Queue.Length())
			for i := 0; i < len(updates); i++ {
				updates[i] = c.Queue.Remove().(update)
			}
		}()

		if len(updates) == 0 {
			continue
		}

		c.Ctx.Debug("sending ws-updates", zap.Int("amount", len(updates)))
		for _, u := range updates {
			err = c.Write(u.Content)
			if err == nil {
				continue
			}

			if websocket.IsUnexpectedCloseError(err) {
				c.Ctx.Info("remote nws client forcely closed connection")
				return
			}

			if _, ok := err.(*net.OpError); ok {
				c.Ctx.Info("error while writing to ws")
				return
			}

			c.Ctx.Warn("unknown error during sending nws update", zap.ByteString("content", u.Content), zap.Error(err))
		}
	}
}
