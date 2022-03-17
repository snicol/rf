package rpc

import (
	"net/http"

	"github.com/snicol/rf"
)

func (rpc *Handler) Handle() rf.HandlerFunc {
	return rpc.fn
}

func result(w http.ResponseWriter, body string, statusCode int, contentType *string) {
	var headerContentType = "text/plain"

	if contentType != nil {
		headerContentType = *contentType
	}

	w.Header().Add("Content-Type", headerContentType)
	w.WriteHeader(statusCode)
	w.Write([]byte(body))
}
