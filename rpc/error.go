package rpc

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/snicol/rf"

	"github.com/snicol/yael"
)

func (rpc *Handler[Req, Res]) Error() rf.ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		yaelErr := &yael.E{}

		ok := errors.As(err, &yaelErr)
		if !ok {
			unknown := yael.New("unknown")
			unknownJSON, _ := json.Marshal(unknown)
			result(w, string(unknownJSON), http.StatusInternalServerError, defaultContentType)
			return
		}

		yaelJSON, err := json.Marshal(yaelErr)
		if err != nil {
			result(w, err.Error(), http.StatusInternalServerError, "")
			return
		}

		result(w, string(yaelJSON), yael.StatusCode(*yaelErr), defaultContentType)
	}
}
