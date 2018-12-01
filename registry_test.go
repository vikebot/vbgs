package main

import "testing"

func TestRegistryClientPut(t *testing.T) {
	registryInit()

	err := ntcpRegistry.Put(&ntcpclient{UserID: 0})
	if err != nil {
		t.Fail()
	}

	err = ntcpRegistry.Put(&ntcpclient{UserID: 0})
	if err != nil {
		if err.Error() != "user(0) already exists in registry 'regntcp'" {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestRegistryClientGet(t *testing.T) {
	registryInit()

	err := ntcpRegistry.Put(&ntcpclient{UserID: 0, Authenticated: true})
	if err != nil {
		t.Fail()
	}

	c := ntcpRegistry.Get(0)
	if c == nil || c.Authenticated != true {
		t.Fail()
	}
}
