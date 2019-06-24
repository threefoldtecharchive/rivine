package api

import "github.com/julienschmidt/httprouter"

// Router represents an http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes.
type Router interface {
	GET(path string, handle httprouter.Handle)
	POST(path string, handle httprouter.Handle)
	OPTIONS(path string, handle httprouter.Handle)
}
