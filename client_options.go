package ario

import "github.com/saeidrp/aria2-rpc/caller"

// DefaultWSReadLimit is the default maximum size in bytes for a single
// WebSocket message.
const DefaultWSReadLimit int64 = caller.DefaultWSReadLimit

// ClientOptions configures Client creation behavior.
type ClientOptions struct {
	// WSReadLimit sets the maximum size in bytes for a single WebSocket message.
	// If zero, DefaultWSReadLimit is used.
	WSReadLimit int64
}

func (o *ClientOptions) toCallerOptions() *caller.Options {
	if o == nil {
		return nil
	}
	return &caller.Options{
		WSReadLimit: o.WSReadLimit,
	}
}
