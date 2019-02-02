package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/vikebot/vbcore"
	"github.com/vikebot/vbgs/vbge"
	"go.uber.org/zap"
)

type ntcpclient struct {
	Out           io.Writer
	LogCtx        *zap.Logger
	Authenticated bool
	UserID        int
	Crypt         *vbcore.CryptoService
	CurType       string
	SDK           string
	SDKLink       string
	OS            string
	Player        *vbge.Player

	IP     string
	PureIP string

	LoginDone       bool
	ClienthelloDone bool
	AgreeconnDone   bool
	IsEncrypted     bool

	StartPc uint32
	Pc      uint32
}

func newNtcpclient(ip net.Addr, w io.Writer, ctx *zap.Logger) *ntcpclient {
	pureip := ip.String()
	if addr, ok := ip.(*net.TCPAddr); ok {
		pureip = addr.IP.String()
	}
	return &ntcpclient{
		LogCtx:  ctx,
		IP:      ip.String(),
		PureIP:  pureip,
		Out:     w,
		CurType: "unknown",
	}
}

type typePacket struct {
	Type *string      `json:"type"`
	Pc   *uint32      `json:"pc"`
	Obj  *interface{} `json:"obj"`
}

type defaultResponse struct {
	Type  string  `json:"type"`
	Pc    *uint32 `json:"pc,omitempty"`
	Error *string `json:"error"`
}

func newDefaultResponse(c *ntcpclient, err *string) defaultResponse {
	dr := defaultResponse{
		Type:  c.CurType,
		Error: err,
	}
	if c.IsEncrypted {
		c.Pc++
		dr.Pc = &c.Pc
	}
	return dr
}

type defaultObjResponse struct {
	Type  string      `json:"type"`
	Pc    *uint32     `json:"pc,omitempty"`
	Error *string     `json:"error"`
	Obj   interface{} `json:"obj"`
}

func newDefaultObjResponse(c *ntcpclient, d interface{}) defaultObjResponse {
	dr := defaultObjResponse{
		Type:  c.CurType,
		Error: nil,
		Obj:   d,
	}
	if c.IsEncrypted {
		c.Pc++
		dr.Pc = &c.Pc
	}
	return dr
}

func (c *ntcpclient) MgmtWrite(d interface{}) {
	buf, err := json.Marshal(d)
	if err != nil {
		c.LogCtx.Warn("failed to marshal interface", zap.Error(err))
		return
	}

	// Precreate debug message (only print it if encryption succeeds)
	pkt := string(buf)

	// Encrypt
	if c.IsEncrypted && !envDisableCrypt {
		cipher, err := c.Crypt.EncryptBase64(buf)
		if err != nil {
			c.LogCtx.Error("encrypting buffer failed", zap.Error(err))
			return
		}
		buf = cipher
	}

	c.LogCtx.Debug("sent",
		zap.String("packet", pkt),
		zap.Uint32("seqnr", c.Pc-c.StartPc))

	buf = append(buf, '\n')
	_, err = c.Out.Write(buf)
	if err != nil {
		c.LogCtx.Warn("sending failed", zap.Error(err))
	}
}

func (c *ntcpclient) RespondNil() {
	c.MgmtWrite(newDefaultResponse(c, nil))
}

func (c *ntcpclient) Respond(errorText string) {
	c.MgmtWrite(newDefaultResponse(c, &errorText))
}

func (c *ntcpclient) RespondFmt(format string, a ...interface{}) {
	c.Respond(fmt.Sprintf(format, a...))
}

func (c *ntcpclient) RespondObj(d interface{}) {
	c.MgmtWrite(newDefaultObjResponse(c, d))
}

func (c *ntcpclient) InitAes(key []byte) error {
	cs, err := vbcore.NewCryptoService(key)
	if err != nil {
		return err
	}

	c.Crypt = cs
	return nil
}
