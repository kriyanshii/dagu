// Code generated by go-swagger; DO NOT EDIT.

package dags

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// SearchDAGsHandlerFunc turns a function with the right signature into a search d a gs handler
type SearchDAGsHandlerFunc func(SearchDAGsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn SearchDAGsHandlerFunc) Handle(params SearchDAGsParams) middleware.Responder {
	return fn(params)
}

// SearchDAGsHandler interface for that can handle valid search d a gs params
type SearchDAGsHandler interface {
	Handle(SearchDAGsParams) middleware.Responder
}

// NewSearchDAGs creates a new http.Handler for the search d a gs operation
func NewSearchDAGs(ctx *middleware.Context, handler SearchDAGsHandler) *SearchDAGs {
	return &SearchDAGs{Context: ctx, Handler: handler}
}

/*
	SearchDAGs swagger:route GET /search dags searchDAGs

# Search DAGs

Searches for DAGs based on a query string.
*/
type SearchDAGs struct {
	Context *middleware.Context
	Handler SearchDAGsHandler
}

func (o *SearchDAGs) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewSearchDAGsParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
