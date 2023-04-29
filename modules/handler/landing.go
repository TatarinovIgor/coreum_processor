package handler

import (
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
)

// PageLanding parses a landing page, set to another routing in AWS to prevent accidental requests
func PageLanding(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	t, err := template.ParseFiles("./templates/landing/index.html")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
		return
	}

}
