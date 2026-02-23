package caller

import (
	"fmt"
	"net/url"
)

// DefaultWSReadLimit is the default maximum size in bytes for a single
// WebSocket message.
const DefaultWSReadLimit int64 = 8 << 20 // 8 MiB

// Options configures transport behavior for caller creation.
type Options struct {
	// WSReadLimit sets the maximum size in bytes for a single WebSocket message.
	// If zero, DefaultWSReadLimit is used.
	WSReadLimit int64
}

func (o *Options) wsReadLimit() int64 {
	if o == nil || o.WSReadLimit == 0 {
		return DefaultWSReadLimit
	}
	return o.WSReadLimit
}

type Caller struct {
	Call  func(method string, params, reply any) error
	Close func() error
}

func NewCaller(host *url.URL) (*Caller, error) {
	return NewCallerWithOptions(host, nil)
}

// NewCallerWithOptions creates a caller for the given host with transport
// options.
func NewCallerWithOptions(host *url.URL, opts *Options) (*Caller, error) {
	rpc := &Caller{}

	switch host.Scheme {
	case "http", "https":
		h, err := newHttpCaller(host.String())
		if err != nil {
			return nil, err
		}

		rpc.Call = h.call
		rpc.Close = h.close
	case "ws", "wss":
		w, err := newWsCaller(host.String(), opts.wsReadLimit())
		if err != nil {
			return nil, err
		}

		rpc.Call = w.call
		rpc.Close = w.close
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", host.Scheme)
	}
	return rpc, nil
}
