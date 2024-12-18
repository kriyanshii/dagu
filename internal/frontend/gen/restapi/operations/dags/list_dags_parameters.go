// Code generated by go-swagger; DO NOT EDIT.

package dags

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NewListDagsParams creates a new ListDagsParams object
//
// There are no default values defined in the spec.
func NewListDagsParams() ListDagsParams {

	return ListDagsParams{}
}

// ListDagsParams contains all the bound params for the list dags operation
// typically these are obtained from a http.Request
//
// swagger:parameters listDags
type ListDagsParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  In: query
	*/
	Limit *int64
	/*
	  In: query
	*/
	Page *int64
	/*
	  In: query
	*/
	SearchName *string
	/*
	  In: query
	*/
	SearchStatus *string
	/*
	  In: query
	*/
	SearchTag *string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewListDagsParams() beforehand.
func (o *ListDagsParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qLimit, qhkLimit, _ := qs.GetOK("limit")
	if err := o.bindLimit(qLimit, qhkLimit, route.Formats); err != nil {
		res = append(res, err)
	}

	qPage, qhkPage, _ := qs.GetOK("page")
	if err := o.bindPage(qPage, qhkPage, route.Formats); err != nil {
		res = append(res, err)
	}

	qSearchName, qhkSearchName, _ := qs.GetOK("searchName")
	if err := o.bindSearchName(qSearchName, qhkSearchName, route.Formats); err != nil {
		res = append(res, err)
	}

	qSearchStatus, qhkSearchStatus, _ := qs.GetOK("searchStatus")
	if err := o.bindSearchStatus(qSearchStatus, qhkSearchStatus, route.Formats); err != nil {
		res = append(res, err)
	}

	qSearchTag, qhkSearchTag, _ := qs.GetOK("searchTag")
	if err := o.bindSearchTag(qSearchTag, qhkSearchTag, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindLimit binds and validates parameter Limit from query.
func (o *ListDagsParams) bindLimit(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("limit", "query", "int64", raw)
	}
	o.Limit = &value

	return nil
}

// bindPage binds and validates parameter Page from query.
func (o *ListDagsParams) bindPage(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("page", "query", "int64", raw)
	}
	o.Page = &value

	return nil
}

// bindSearchName binds and validates parameter SearchName from query.
func (o *ListDagsParams) bindSearchName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.SearchName = &raw

	return nil
}

// bindSearchStatus binds and validates parameter SearchStatus from query.
func (o *ListDagsParams) bindSearchStatus(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.SearchStatus = &raw

	return nil
}

// bindSearchTag binds and validates parameter SearchTag from query.
func (o *ListDagsParams) bindSearchTag(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.SearchTag = &raw

	return nil
}
