package ario

import (
	"context"
	"io"
	"log"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

type reply struct {
	Method string  `json:"method"`
	Params []Event `json:"params"`
}

type Event struct {
	Gid string `json:"gid"`
}

type Events = map[string]func(gid string)

type notifier struct {
	conn *websocket.Conn
	mux  sync.Mutex
}

type Notify struct {
	r     map[string]chan string
	Close func() error
	mux   *sync.Mutex
}

const (
	ods = "aria2.onDownloadStart"
	odp = "aria2.onDownloadPause"
	odt = "aria2.onDownloadStop"
	odc = "aria2.onDownloadComplete"
	ode = "aria2.onDownloadError"
	obc = "aria2.onBtDownloadComplete"
)

func newNotifier(host *url.URL) (*notifier, error) {
	switch host.Scheme {
	case "https", "wss":
		host.Scheme = "wss"
	case "http", "ws":
		host.Scheme = "ws"
	}

	conn, _, err := websocket.DefaultDialer.Dial(host.String(), nil)
	if err != nil {
		return nil, err
	}

	return &notifier{
		conn: conn,
	}, nil
}

func (n *notifier) Listener(ctx context.Context) (*Notify, error) {
	r := make(map[string]chan string)

	go func() {
		defer func() {
			for _, v := range r {
				close(v)
			}
			n.conn.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			resp := &reply{}

			// read notifications from the connection
			if err := n.conn.ReadJSON(resp); err != nil {
				if err == io.ErrUnexpectedEOF {
					log.Println("unexpected EOF | if you are using nginx, please adjust the value of `proxy_read_timeout`")
					return
				}
				log.Printf("reading websocket message: %v", err)
				return
			}

			for _, event := range resp.Params {
				// only send when the channel exists
				n.mux.Lock()
				ch, ok := r[resp.Method]
				n.mux.Unlock()

				if ok {
					ch <- event.Gid
				}
			}
		}
	}()

	return &Notify{
		r,
		n.conn.Close,
		&n.mux,
	}, nil
}

func (n *Notify) notifyFunc(method string) <-chan string {
	n.mux.Lock()
	gid, ok := n.r[method]
	n.mux.Unlock()

	if ok {
		return gid
	}

	// if channel not exist, create it
	n.mux.Lock()
	gid = make(chan string)
	n.r[method] = gid
	n.mux.Unlock()

	return gid
}

func (n *Notify) MultiListen(events map[string]func(gid string)) {
	for method, fn := range events {
		go func(m string, f func(gid string)) {
			for gid := range n.notifyFunc(m) {
				f(gid)
			}
		}(method, fn)
	}
}

func (n *Notify) Start() <-chan string {
	return n.notifyFunc(ods)
}

func (n *Notify) Pause() <-chan string {
	return n.notifyFunc(odp)
}

func (n *Notify) Stop() <-chan string {
	return n.notifyFunc(odt)
}

func (n *Notify) Complete() <-chan string {
	return n.notifyFunc(odc)
}

func (n *Notify) Error() <-chan string {
	return n.notifyFunc(ode)
}

func (n *Notify) BtComplete() <-chan string {
	return n.notifyFunc(obc)
}
