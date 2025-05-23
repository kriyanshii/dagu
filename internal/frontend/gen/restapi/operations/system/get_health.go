// Code generated by go-swagger; DO NOT EDIT.

package system

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetHealthHandlerFunc turns a function with the right signature into a get health handler
type GetHealthHandlerFunc func(GetHealthParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetHealthHandlerFunc) Handle(params GetHealthParams) middleware.Responder {
	return fn(params)
}

// GetHealthHandler interface for that can handle valid get health params
type GetHealthHandler interface {
	Handle(GetHealthParams) middleware.Responder
}

// NewGetHealth creates a new http.Handler for the get health operation
func NewGetHealth(ctx *middleware.Context, handler GetHealthHandler) *GetHealth {
	return &GetHealth{Context: ctx, Handler: handler}
}

/*
	GetHealth swagger:route GET /health system getHealth

# Health check endpoint

Returns the health status of the server and its dependencies
*/
type GetHealth struct {
	Context *middleware.Context
	Handler GetHealthHandler
}

func (o *GetHealth) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetHealthParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
