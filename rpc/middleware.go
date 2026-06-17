package rpc

import (
	"github.com/snicol/rf"
	"github.com/snicol/rf/middleware"
)

// DefaultMiddleware returns the default middleware stack for RPC handlers.
func DefaultMiddleware() []rf.MiddlewareFunc {
	return []rf.MiddlewareFunc{
		middleware.RPCRequestOnly(),
	}
}
