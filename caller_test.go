package ario

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/saeidrp/aria2-rpc/caller"
	"github.com/saeidrp/aria2-rpc/testutils"
)

func TestHTTPRPC(t *testing.T) {
	uri, _ := url.Parse(testutils.Arai2Uri("https://"))
	c, err := caller.NewCaller(uri)
	if err != nil {
		fmt.Println(err)
		t.Fatal("NewCaller should not return error")
	}

	t.Run("connect should not be error", func(t *testing.T) {
		r := Version{}
		err = c.Call("aria2.getVersion", nil, &r)
		if err != nil {
			t.Fatal("get version failed: ", err)
		}

		t.Log(r)
	})

	t.Run("when reply is nil, error should not be returned.", func(t *testing.T) {
		err = c.Call("aria2.getVersion", nil, nil)
		if err != nil {
			t.Fatal("get version failed: ", err)
		}
	})
}

func TestWSRPC(t *testing.T) {
	uri, _ := url.Parse(testutils.Arai2Uri("wss://"))
	c, err := caller.NewCaller(uri)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("connect should not be error", func(t *testing.T) {
		r := Version{}
		err = c.Call("aria2.getVersion", nil, &r)
		if err != nil {
			t.Fatal("get version failed: ", err)
		}

		t.Log(r)
	})

	t.Run("when reply is nil, error should not be returned.", func(t *testing.T) {
		err = c.Call("aria2.getVersion", nil, nil)
		if err != nil {
			t.Fatal("get version failed: ", err)
		}
	})
}
