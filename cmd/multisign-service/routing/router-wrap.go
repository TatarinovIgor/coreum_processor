package routing

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// RouterWrap is proxy over httprouter.Router to allow setup CORS in auto-mode
type RouterWrap struct {
	prefix     string
	router     *httprouter.Router
	optionUrls map[string]bool
}

// GET is wrap method over adding GET method handler and OPTION handler to allow CORS
func (r *RouterWrap) GET(path string, handle httprouter.Handle) {
	r.router.GET(r.prefix+path, handle)
	r.OptionsCors(path, optionsCORSHandler)
}

// POST is wrap method over adding POST method handler and OPTION handler to allow CORS
func (r *RouterWrap) POST(path string, handle httprouter.Handle) {
	r.router.POST(r.prefix+path, handle)
	r.OptionsCors(path, optionsCORSHandler)
}

// PUT is wrap method over adding PUT method handler and OPTION handler to allow CORS
func (r *RouterWrap) PUT(path string, handle httprouter.Handle) {
	r.router.PUT(r.prefix+path, handle)
	r.OptionsCors(path, optionsCORSHandler)
}

// DELETE is wrap method over adding DELETE method handler and OPTION handler to allow CORS
func (r *RouterWrap) DELETE(path string, handle httprouter.Handle) {
	r.router.DELETE(r.prefix+path, handle)
	r.OptionsCors(path, optionsCORSHandler)
}

// OptionsCors controls duplicates for OPTION urls
func (r *RouterWrap) OptionsCors(path string, handle httprouter.Handle) {
	if _, ok := r.optionUrls[r.prefix+path]; !ok {
		r.router.OPTIONS(r.prefix+path, optionsCORSHandler)
		r.optionUrls[r.prefix+path] = true
	}
}

// NewRouterWrap is constructor for RouterWrap
func NewRouterWrap(prefix string, router *httprouter.Router) *RouterWrap {
	return &RouterWrap{
		prefix:     prefix,
		router:     router,
		optionUrls: map[string]bool{},
	}
}

func optionsCORSHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json") // normal header
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Authorization, origin, x-requested-with, content-type")
}
