package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eapache/queue"
	"github.com/vikebot/vbdb"
	"github.com/vikebot/vbgs/vbge"
	"go.uber.org/zap"
)

const (
	notifyChannelPrivate   = 0
	notifyChannelGroup     = 1
	notifyChannelBroadcast = 2
)

const (
	severityDefault = 0
	severitySuccess = 1
	severityWarning = 2
	severityError   = 3
)

type update struct {
	UnixN   int64
	Content []byte
}

func newUpdate(t string, buf []byte) update {
	u := update{
		UnixN: time.Now().UTC().UnixNano(),
	}
	u.Content = []byte(`{"type":"` + t + `","obj":[` + string(buf) + `],"unixn":` + strconv.FormatInt(u.UnixN, 10) + `}`)
	return u
}

type updateDistributor struct {
	History         *queue.Queue
	HistorySyncRoot sync.Mutex
}

func (ud *updateDistributor) Push(sender *vbge.Player, u update, notifyChan int, ng vbge.NotifyGroup, ctx *zap.Logger) {
	ud.HistoryAdd(u)

	switch notifyChan {
	case notifyChannelPrivate:
		ud.Notify(sender.UserID, u, ctx)
	case notifyChannelGroup:
		for _, groupMember := range ng {
			ud.Notify(groupMember.UserID, u, ctx)
		}
	case notifyChannelBroadcast:
		// Anonym function in order to use defered .Unlock() statement and
		// hence insure a unlock even after a eventual panic
		func() {
			nwsRegistry.baton.Lock()
			defer nwsRegistry.baton.Unlock()

			for _, clients := range nwsRegistry.m {
				for _, c := range clients {
					c.Notify(u)
				}
			}
		}()
	}
}

func (ud *updateDistributor) Notify(userID int, u update, ctx *zap.Logger) {
	clients := nwsRegistry.Get(userID)
	for _, c := range clients {
		c.Notify(u)
	}
}

func (ud *updateDistributor) WriteChat(sender *vbge.Player, msg string, severity int, notifyChan int, ng vbge.NotifyGroup, ctx *zap.Logger) {
	ud.WriteChatPrefixed(sender, "[Server] ", msg, severity, notifyChan, ng, ctx)
}

func (ud *updateDistributor) WriteChatPrefixed(sender *vbge.Player, prefix string, msg string, severity int, notifyChan int, ng vbge.NotifyGroup, ctx *zap.Logger) {
	var class string
	switch severity {
	case severityDefault:
		class = ""
	case severitySuccess:
		class = "green"
	case severityWarning:
		class = "yellow"
	case severityError:
		class = "red"
	}

	elem := "<p class=\"" + class + "\">" + prefix + msg + "</p>&nbsp;"
	ud.Push(sender, newUpdate("chat", []byte("\""+strings.Replace(elem, "\"", "\\\"", -1)+"\"")), notifyChan, ng, ctx)
}

func (ud *updateDistributor) PushTypeUserinfo(ws *nwsclient) {
	var player *vbge.Player
	var ok bool
	if player, ok = battle.Players[ws.UserID]; !ok {
		ws.Log.Warn("unable to find player for connected (legit) websocket")
		return
	}

	user, success := vbdb.UserFromIDCtx(ws.UserID, ws.Log)
	if !success {
		ws.Log.Warn("unable to load user")
		return
	}

	u := update{
		UnixN: time.Now().UnixNano(),
	}

	u.Content = []byte(fmt.Sprintf(`{"type":"userinfo","obj":{"user":{"name":"%s","username":"%s","picture":"%s","permission":"%s"}},"unixn":%s}`,
		user.Name,
		user.Username,
		config.Battle.AvatarPictureURL+player.PicLink,
		user.PermissionString,
		strconv.FormatInt(u.UnixN, 10)))

	ws.Notify(u)
}

func (ud *updateDistributor) PushTypeInfo(c *ntcpclient, established bool) {
	clients := nwsRegistry.Get(c.UserID)
	if len(clients) == 0 {
		return
	}

	u := update{
		UnixN: time.Now().UnixNano(),
	}
	u.Content = []byte(fmt.Sprintf(`{"type":"info","obj":{"conn":{"established":%s,"ip":"%s"},"lib":{"name":"%s","link":"%s"},"os":"%s"},"unixn":%s}`,
		strconv.FormatBool(established),
		c.PureIP,
		c.SDK,
		c.SDKLink,
		c.OS,
		strconv.FormatInt(u.UnixN, 10)))

	for _, ws := range clients {
		ws.Notify(u)
	}
}

func (ud *updateDistributor) PushTypeFlag(c *nwsclient, name string, state bool) {
	u := update{
		UnixN: time.Now().UnixNano(),
	}
	u.Content = []byte(fmt.Sprintf(`{"type":"flag","obj":{"name":"%s","state":%s},"unixn":%s}`, name, strconv.FormatBool(state), strconv.FormatInt(u.UnixN, 10)))

	c.Notify(u)
}

func (ud *updateDistributor) PushInit(c *nwsclient, obj []byte) {
	u := update{
		UnixN: time.Now().UnixNano(),
	}

	u.Content = []byte(fmt.Sprintf(`{"type":"initial","obj":%s,"unixn":%s}`, obj, strconv.FormatInt(u.UnixN, 10)))

	c.Notify(u)
}

func (ud *updateDistributor) PushStats(c *nwsclient, obj []byte) {
	u := update{
		UnixN: time.Now().UnixNano(),
	}

	u.Content = []byte(fmt.Sprintf(`{"type":"stats","obj":[%s],"unixn":%s}`, obj, strconv.FormatInt(u.UnixN, 10)))

	c.Notify(u)
}

func (ud *updateDistributor) HistoryAdd(u update) {
	ud.HistorySyncRoot.Lock()
	defer ud.HistorySyncRoot.Unlock()

	ud.History.Add(u)
}

func (ud *updateDistributor) InitHistoryWorker(start chan bool, shutdown chan bool) error {
	fn := "vb_" + config.Instance + "_history.log"
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	go func() {
		defer f.Close()
		ctx := log.With(zap.String("worker", "history_appender"))

		var totsafed int64
		totsafed = 0

		<-start
		for {
			time.Sleep(time.Second * 5)

			shouldExit := false
			select {
			case <-shutdown:
				ctx.Info("received shutdown signal. checking last time for updates")
				shouldExit = true
			default:
			}

			var updates []update
			func() {
				ud.HistorySyncRoot.Lock()
				defer ud.HistorySyncRoot.Unlock()

				updates = make([]update, ud.History.Length())
				for i := 0; i < len(updates); i++ {
					updates[i] = ud.History.Remove().(update)
				}
			}()

			if len(updates) > 0 {
				totsafed += int64(len(updates))
				ctx.Info("appending updates to log file", zap.Int("amount", len(updates)))
				for _, item := range updates {
					f.WriteString(strconv.FormatInt(item.UnixN, 10) + ";" + base64.StdEncoding.EncodeToString(item.Content) + "\n")
				}
				ctx.Info("finished appending. sleeping till next round")
			}

			if shouldExit {
				ctx.Info("stopping worker. exiting", zap.Int64("total_appended_updates", totsafed))
				return
			}
		}
	}()

	return nil
}
