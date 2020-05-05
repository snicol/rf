package rpc

import (
	"encoding/json"
	"net/http"

	"github.com/snicol/rf"

	"github.com/snicol/yael"
)

func (rpc *Handler) Error() rf.ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		var v interface{} = err
		yaelErr, ok := v.(*yael.E)
		if !ok {
			result(w, "unknown error", http.StatusInternalServerError, nil)
			return
		}

		yaelJSON, err := json.Marshal(yaelErr)
		if err != nil {
			result(w, err.Error(), http.StatusInternalServerError, nil)
			return
		}

		result(w, string(yaelJSON), yael.StatusCode(*yaelErr), &defaultContentType)
	}
}
