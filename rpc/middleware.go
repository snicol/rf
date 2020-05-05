package rpc

import (
	"github.com/snicol/rf"
	"github.com/snicol/rf/middleware"
)

func DefaultMiddleware() []rf.MiddlewareFunc {
	return []rf.MiddlewareFunc{
		middleware.RPCRequestOnly(),
	}
}
