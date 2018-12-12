package main

import (
	"net"
	"net/http"
	"net/url"
	"time"

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
		log.Info("nws ready. waiting for start signal")
		<-start

		go nwsRun(srv)

		// Shutdown websocket when signal is received
		<-shutdown
		err := srv.Shutdown(nil)
		if err != nil {
			log.Warn("nws shutdown failed", zap.Error(err))
		}
	}()
}

func nwsRun(srv *http.Server) {
	var srvErr error

	log.Info("accepting clients on nws listener")
	if config.Network.WS.TLS.Active {
		srvErr = srv.ListenAndServeTLS(config.Network.WS.TLS.Cert, config.Network.WS.TLS.PKey)
	} else {
		srvErr = srv.ListenAndServe()
	}

	if srvErr != nil {
		log.Fatal("nws listen failed", zap.Error(srvErr))
	}
}

func nwsHandler(w http.ResponseWriter, r *http.Request) {
	wsrqid := vbcore.FastRandomString(32)
	c := &nwsclient{
		WSRqID: wsrqid,
		Log:    log.With(zap.String("wsrqid", wsrqid)),
	}

	c.Log.Info("connected", zap.String("ip", r.RemoteAddr))

	ws, err := nwsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		c.Log.Error("failed to upgrade http connection", zap.Error(err))
		return
	}
	defer func() {
		c.Log.Info("closed", zap.String("ip", r.RemoteAddr))
		err = ws.Close()
		if err != nil {
			c.Log.Error("error during closing websocket", zap.Error(err))
		}
	}()

	c.Ws = ws

	nws(c)
}

func nws(c *nwsclient) {
	mt, watchtoken, err := c.Ws.ReadMessage()
	if err != nil {
		c.Log.Warn("failed reading message from websocket", zap.Error(err))
		return
	}
	c.Mt = mt

	v, exists, success := vbdb.RoundentryFromWatchtokenCtx(string(watchtoken), c.Log)
	if !success {
		err = c.WriteStr("Internal server error")
		if err != nil {
			c.Log.Warn("unable to send internal server error message", zap.Error(err))
		}
		return
	}
	if !exists {
		c.Log.Warn("client provided unknown watchtoken", zap.String("watchtoken", string(watchtoken)))
		err = c.WriteStr("Unknown watchtoken")
		if err != nil {
			c.Log.Warn("unable to send unknown watchtoken error message", zap.Error(err))
		}
		return
	}
	c.UserID = v.UserID
	c.Log.Info("authenticated", zap.Int("user_id", v.UserID))

	if config.Battle.RoundID != v.RoundID {
		c.Log.Warn("valid watchtoken references invalid round",
			zap.Int("config_round_id", config.Battle.RoundID),
			zap.Int("watchtoken_round_id", v.RoundID))

		err = c.WriteStr("Internal server error")
		if err != nil {
			c.Log.Warn("unable to send internal server error message", zap.Error(err))
		}
		return
	}

	go func() {
		time.Sleep(1 * time.Second)

		// send user info
		// updateDist.PushTypeUserinfo(c)

		var player = battle.Players[c.UserID]

		viewableMapsize := vbge.Location{
			X: vbge.RenderWidth,
			Y: vbge.RenderHeight,
		}

		playerMapentity, err := vbge.GetViewableMapentity(viewableMapsize.X, viewableMapsize.Y, c.UserID, battle, true)
		if err != nil {
			c.Log.Error("failed getting mapentity", zap.Error(err))
			return
		}

		dist.GetClient(c.UserID).Push("initial",
			struct {
				TotalMapsize    vbge.Location        `json:"totalmapsize"`
				ViewableMapsize vbge.Location        `json:"viewablemapsize"`
				MaxHealth       int                  `json:"maxhealth"`
				PlayerMapentity [][]*vbge.EntityResp `json:"playermapentity"`
				Startplayer     string               `json:"startplayer"`
			}{
				TotalMapsize: vbge.Location{
					X: vbge.MapWidth,
					Y: vbge.MapHeight,
				},
				ViewableMapsize: viewableMapsize,
				MaxHealth:       vbge.MaxHealth,
				Startplayer:     player.GRenderID,
				PlayerMapentity: playerMapentity.Matrix,
			},
			c.Log)

		c.Log.Debug("sending init package to nwsclient")
	}()

	// subscribe websocket connection for all notifications to this user and
	// send them as long as err isn't a disconnect from the remote websocket
	dist.GetClient(c.UserID).Sub(func(notf []byte) (disconnected bool, err error) {
		err = c.Write(notf)
		if err == nil {
			return
		}

		if _, ok := err.(*net.OpError); ok || websocket.IsUnexpectedCloseError(err) {
			return true, err
		}

		return
	}, c.Log)

	if config.Network.WS.Flags.Debug {

		dist.GetClient(c.UserID).Push("flag", struct {
			Name  string `json:"name"`
			State bool   `json:"state"`
		}{
			"debug",
			true,
		}, c.Log)
		c.Log.Debug("sending debug flag to nwsclient")
	}

	if config.Network.WS.Flags.Stats {
		// start goroutinge because pushStats can block the
		// init packet if it's taken very long
		go pushStats(c)
	}
}

func pushStats(c *nwsclient) {
	stats, err := getPlayersStats()
	if err != nil {
		c.Log.Error("failed getting stats", zap.Error(err))
		return
	}

	dist.GetClient(c.UserID).Push("stats", struct {
		Stats playersStats
	}{
		stats,
	}, c.Log)

	c.Log.Debug("sending stats package to nwsclient")
}
