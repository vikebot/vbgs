package main

import (
	"bytes"
	"testing"
)

type packetTest struct {
	In           string
	Out          string
	ActualOutput []byte
	PacketType   string
	Client       *ntcpclient
	Mf           func(c *ntcpclient)
}

func newPT(in string, out string, packetType string) packetTest {
	p := packetTest{
		In:         in,
		Out:        out,
		PacketType: packetType,
	}
	return p
}

func newPTMf(in string, out string, packetType string, mf func(c *ntcpclient)) packetTest {
	p := packetTest{
		In:         in,
		Out:        out,
		PacketType: packetType,
		Mf:         mf,
	}
	return p
}

func (p *packetTest) CallHandler() {
	buf := bytes.NewBuffer(nil)
	c := ntcpclient{
		Out:    buf,
		LogCtx: logctx,
	}
	if p.Mf != nil {
		p.Mf(&c)
	}
	p.Client = &c
	packetHandler(p.Client, []byte(p.In))
	p.ActualOutput = buf.Bytes()
}

func (p *packetTest) AssertOutput() bool {
	if len(p.ActualOutput) > 0 {
		aoStr := string(p.ActualOutput[:len(p.ActualOutput)-1])
		return aoStr == p.Out
	}
	return p.Out == ""
}

func (p *packetTest) AssertPacketType() bool {
	return p.Client.CurType == p.PacketType
}

func TestPacketJSONSyntax(t *testing.T) {
	output := `{"type":"forbidden","error":"Invalid JSON syntax"}`
	cases := []packetTest{
		newPT("", output, ""),
		newPT(`{""}`, output, ""),
		newPT(`{dasfasdf}`, output, ""),
		newPT(`1586496`, output, ""),
		newPT(`{"type":"login","obj":}`, output, ""),
		newPT(`{"type:clienthello","obj":{}}`, output, ""),
		newPT(`{"type":login,"obj":{}}`, output, ""),
		newPT(`{"type":true,"obj":{}}`, output, ""),
		newPT(`"type":"true","obj":{}`, output, ""),
	}

	for _, v := range cases {
		v.CallHandler()
		if !v.AssertOutput() {
			t.Fail()
		}
	}
}

func TestPacketDefaultTypes(t *testing.T) {
	cases := []packetTest{
		newPT(`{}`,
			`{"type":"forbidden","error":"Invalid packet. '.type' missing"}`, ""),
		newPT(`{"obj":{}}`,
			`{"type":"forbidden","error":"Invalid packet. '.type' missing"}`, ""),
	}

	for _, v := range cases {
		v.CallHandler()
		if !v.AssertOutput() {
			t.Fail()
		}
	}
}

func TestPacketAlreadyDone(t *testing.T) {
	cases := []packetTest{
		newPTMf(`{"type":"login","obj":{"roundticket":"xxx"}}`,
			`{"type":"login","error":"Protocol mismatch. 'login' already done"}`, "", func(c *ntcpclient) {
				c.LoginDone = true
			}),
		newPTMf(`{"type":"clienthello","obj":{"cipher":"xxx"}}`,
			`{"type":"clienthello","error":"Protocol mismatch. 'clienthello' already done"}`, "", func(c *ntcpclient) {
				c.ClienthelloDone = true
			}),
		newPTMf(`{"type":"agreeconn","obj":{"cipher":"xxx"}}`,
			`{"type":"agreeconn","error":"Protocol mismatch. 'agreeconn' already done"}`, "", func(c *ntcpclient) {
				c.AgreeconnDone = true
			}),
	}

	for _, v := range cases {
		v.CallHandler()
		if !v.AssertOutput() {
			t.Fail()
		}
	}
}

func TestPacketChronology(t *testing.T) {
	cases := []packetTest{}

	for _, v := range cases {
		v.CallHandler()
		if !v.AssertOutput() {
			t.Fail()
		}
	}
}

func TestPacketTypeDetection(t *testing.T) {
	cases := []packetTest{
		newPT(`{"type":"login","obj":{}}`, "", "login"),
		newPT(`{"type":"clienthello","obj":{}}`, "", "clienthello"),
	}

	for _, v := range cases {
		v.CallHandler()
		if !v.AssertPacketType() {
			t.Fail()
		}
	}
}
