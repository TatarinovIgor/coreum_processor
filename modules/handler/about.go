package handler

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// About is a function which should be used for pinging the server and to check the status of the process
func About(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Write([]byte("about"))
}
