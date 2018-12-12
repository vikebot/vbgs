package main

import (
	"net"
	"net/http"
	"net/url"

	"github.com/vikebot/vbgs/pkg/ntfydistr"

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

	// authenticate the websocket connection
	err = nwsAuthAndValidate(c)
	if err != nil {
		// see if the error happend due to a closed websocket
		if _, ok := err.(*net.OpError); ok || websocket.IsUnexpectedCloseError(err) {
			return
		}

		// no closing error -> log it
		c.Log.Warn("unable to send message to client", zap.Error(err))
	}
}

func nwsAuthAndValidate(c *nwsclient) error {
	// get opening message (should be the watchtoken) from the client
	mt, watchtoken, err := c.Ws.ReadMessage()
	if err != nil {
		c.Log.Warn("failed reading message from websocket", zap.Error(err))
		return nil
	}
	c.Mt = mt

	watchtokenStr := string(watchtoken) // TODO: check for legitimacy of watchtoken

	// check if the watchtoken exists inside the database
	v, exists, success := vbdb.RoundentryFromWatchtokenCtx(watchtokenStr, c.Log)
	if !success {
		return c.WriteStr("Internal server error")
	}
	if !exists {
		c.Log.Warn("client provided unknown watchtoken", zap.String("watchtoken", string(watchtoken)))
		return c.WriteStr("Unknown watchtoken")
	}

	// user is authenticated correctly
	c.UserID = v.UserID
	c.Log.Info("authenticated", zap.Int("user_id", v.UserID))

	// check if the user's watchtoken was intended for this round
	if config.Battle.RoundID != v.RoundID {
		c.Log.Warn("valid watchtoken references invalid round",
			zap.Int("config_round_id", config.Battle.RoundID),
			zap.Int("watchtoken_round_id", v.RoundID))

		return c.WriteStr("Unexpected watchtoken. Maybe your round is already over?")
	}

	// subscribe authenticated client
	nwsSub(c)
	return nil
}

func nwsSub(c *nwsclient) {
	// subscribe websocket connection for all notifications to this user and
	// send them as long as err isn't a disconnect from the remote websocket.
	// Also send all initial informations needed by this specific subscriber,
	// as map properties, etc.
	dist.GetClient(c.UserID).Sub(func(notf []byte) (disconnected bool, err error) {
		// write provided notification to the websocket connection
		err = c.Write(notf)
		if err == nil {
			return
		}

		// check if the remote party disconnected
		if _, ok := err.(*net.OpError); ok || websocket.IsUnexpectedCloseError(err) {
			return true, err
		}

		// unknown error -> just return it
		return false, err
	}, func(initClient ntfydistr.Client) {
		// Start the initialization of the current subscriber
		c.Log.Debug("initing nwsclient subscription for user")

		// Construct necessary primitives
		var player = battle.Players[initClient.UserID()]
		viewableMapsize := vbge.Location{
			X: vbge.RenderWidth,
			Y: vbge.RenderHeight,
		}
		playerMapentity, err := vbge.GetViewableMapentity(viewableMapsize.X, viewableMapsize.Y, initClient.UserID(), battle, true)
		if err != nil {
			c.Log.Error("failed getting mapentity", zap.Error(err))
			return
		}

		// Send the initial game information
		c.Log.Debug("sending init package to nwsclient")
		initClient.Push("initial", struct {
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
		}, c.Log)

		// Set the client's debug flag
		c.Log.Debug("sending debug flag to nwsclient", zap.Bool("debug", config.Network.WS.Flags.Debug))
		initClient.Push("flag", struct {
			Name  string `json:"name"`
			State bool   `json:"state"`
		}{
			"debug",
			config.Network.WS.Flags.Debug,
		}, c.Log)

		// Send the current state fo the stats
		if config.Network.WS.Flags.Stats {
			c.Log.Debug("sending stats to nwsclient")

			stats, err := getPlayersStats()
			if err != nil {
				c.Log.Error("failed getting stats", zap.Error(err))
				return
			}

			initClient.Push("stats", struct {
				Stats playersStats `json:"stats"`
			}{
				stats,
			}, c.Log)
		}
	}, c.Log)
}
