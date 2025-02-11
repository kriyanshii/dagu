// Code generated by go-swagger; DO NOT EDIT.

package dags

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// CreateDAGHandlerFunc turns a function with the right signature into a create d a g handler
type CreateDAGHandlerFunc func(CreateDAGParams) middleware.Responder

// Handle executing the request and returning a response
func (fn CreateDAGHandlerFunc) Handle(params CreateDAGParams) middleware.Responder {
	return fn(params)
}

// CreateDAGHandler interface for that can handle valid create d a g params
type CreateDAGHandler interface {
	Handle(CreateDAGParams) middleware.Responder
}

// NewCreateDAG creates a new http.Handler for the create d a g operation
func NewCreateDAG(ctx *middleware.Context, handler CreateDAGHandler) *CreateDAG {
	return &CreateDAG{Context: ctx, Handler: handler}
}

/*
	CreateDAG swagger:route POST /dags dags createDAG

# Create a new DAG

Creates a new DAG.
*/
type CreateDAG struct {
	Context *middleware.Context
	Handler CreateDAGHandler
}

func (o *CreateDAG) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewCreateDAGParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
