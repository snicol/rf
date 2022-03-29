package rpc

import (
	"encoding/json"
	"net/http"

	"github.com/snicol/rf"

	"github.com/snicol/yael"
)

func (rpc *Handler[Req, Res]) Error() rf.ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		yaelErr, ok := err.(*yael.E)
		if !ok {
			yaelErr = yael.New("unknown")
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
