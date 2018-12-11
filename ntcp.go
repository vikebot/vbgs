package main

import (
	"bufio"
	"net"
	"strings"

	"go.uber.org/zap"
)

func ntcpInit(start chan bool, shutdown chan bool) {
	listener, err := net.Listen("tcp", config.Network.TCP.Addr)
	if err != nil {
		log.Fatal("ntcp listen failed", zap.String("addr", config.Network.TCP.Addr), zap.Error(err))
	}

	go func() {
		// Wait for start signal
		log.Info("ntcp ready. waiting for start signal")
		<-start

		go ntcpRun(listener)

		// Shutdown listener as soon as we get signal from master
		<-shutdown
		err = listener.Close()
		if err != nil {
			log.Warn("ntcp close failed", zap.Error(err))
		}
	}()
}

func ntcpRun(listener net.Listener) {
	log.Info("accepting clients on ntcp listener")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Warn("ntcp accept failed", zap.Error(err))
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			ctx := log.With(zap.String("ip", c.RemoteAddr().String()))

			defer func() {
				recoverd := recover()
				switch rval := recoverd.(type) {
				case nil:
					return
				case error:
					ctx.Error("recoverd from panic",
						zap.Error(rval),
						zap.Stack("recoverd_stack"))
				default:
					ctx.Error("recoverd from panic", zap.Any("unknown_err", rval))
				}
			}()

			ntcp(c, ctx)
		}(conn)
	}
}

func ntcp(conn net.Conn, ctx *zap.Logger) {
	c := newNtcpclient(conn.RemoteAddr(), conn, ctx)
	buf := bufio.NewReader(conn)

	c.Log.Info("connected")

	for {
		data, err := buf.ReadBytes('\n')
		if err != nil {
			if strings.HasSuffix(err.Error(), "An existing connection was forcibly closed by the remote host.") || err.Error() == "EOF" {
				disconnect(c)
				return
			}

			if _, ok := err.(*net.OpError); ok {
				disconnect(c)
				return
			}

			c.Log.Warn("unknown error during ntcp read", zap.Error(err))
			return
		}

		packetHandler(c, data[:len(data)-1])
	}
}

func disconnect(c *ntcpclient) {
	c.Log.Info("disconnected")
	updateDist.PushTypeInfo(c, false)
	ntcpRegistry.Delete(c)
}
