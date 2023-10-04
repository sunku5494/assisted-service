// Code generated by go-swagger; DO NOT EDIT.

package installer

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewV2ResetHostValidationParams creates a new V2ResetHostValidationParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewV2ResetHostValidationParams() *V2ResetHostValidationParams {
	return &V2ResetHostValidationParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewV2ResetHostValidationParamsWithTimeout creates a new V2ResetHostValidationParams object
// with the ability to set a timeout on a request.
func NewV2ResetHostValidationParamsWithTimeout(timeout time.Duration) *V2ResetHostValidationParams {
	return &V2ResetHostValidationParams{
		timeout: timeout,
	}
}

// NewV2ResetHostValidationParamsWithContext creates a new V2ResetHostValidationParams object
// with the ability to set a context for a request.
func NewV2ResetHostValidationParamsWithContext(ctx context.Context) *V2ResetHostValidationParams {
	return &V2ResetHostValidationParams{
		Context: ctx,
	}
}

// NewV2ResetHostValidationParamsWithHTTPClient creates a new V2ResetHostValidationParams object
// with the ability to set a custom HTTPClient for a request.
func NewV2ResetHostValidationParamsWithHTTPClient(client *http.Client) *V2ResetHostValidationParams {
	return &V2ResetHostValidationParams{
		HTTPClient: client,
	}
}

/*
V2ResetHostValidationParams contains all the parameters to send to the API endpoint

	for the v2 reset host validation operation.

	Typically these are written to a http.Request.
*/
type V2ResetHostValidationParams struct {

	/* HostID.

	   The host that its validation is being reset.

	   Format: uuid
	*/
	HostID strfmt.UUID

	/* InfraEnvID.

	   The infra-env of the host that its validation is being reset.

	   Format: uuid
	*/
	InfraEnvID strfmt.UUID

	/* ValidationID.

	   The id of the validation being reset.
	*/
	ValidationID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the v2 reset host validation params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *V2ResetHostValidationParams) WithDefaults() *V2ResetHostValidationParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the v2 reset host validation params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *V2ResetHostValidationParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the v2 reset host validation params
func (o *V2ResetHostValidationParams) WithTimeout(timeout time.Duration) *V2ResetHostValidationParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the v2 reset host validation params
func (o *V2ResetHostValidationParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the v2 reset host validation params
func (o *V2ResetHostValidationParams) WithContext(ctx context.Context) *V2ResetHostValidationParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the v2 reset host validation params
func (o *V2ResetHostValidationParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the v2 reset host validation params
func (o *V2ResetHostValidationParams) WithHTTPClient(client *http.Client) *V2ResetHostValidationParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the v2 reset host validation params
func (o *V2ResetHostValidationParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithHostID adds the hostID to the v2 reset host validation params
func (o *V2ResetHostValidationParams) WithHostID(hostID strfmt.UUID) *V2ResetHostValidationParams {
	o.SetHostID(hostID)
	return o
}

// SetHostID adds the hostId to the v2 reset host validation params
func (o *V2ResetHostValidationParams) SetHostID(hostID strfmt.UUID) {
	o.HostID = hostID
}

// WithInfraEnvID adds the infraEnvID to the v2 reset host validation params
func (o *V2ResetHostValidationParams) WithInfraEnvID(infraEnvID strfmt.UUID) *V2ResetHostValidationParams {
	o.SetInfraEnvID(infraEnvID)
	return o
}

// SetInfraEnvID adds the infraEnvId to the v2 reset host validation params
func (o *V2ResetHostValidationParams) SetInfraEnvID(infraEnvID strfmt.UUID) {
	o.InfraEnvID = infraEnvID
}

// WithValidationID adds the validationID to the v2 reset host validation params
func (o *V2ResetHostValidationParams) WithValidationID(validationID string) *V2ResetHostValidationParams {
	o.SetValidationID(validationID)
	return o
}

// SetValidationID adds the validationId to the v2 reset host validation params
func (o *V2ResetHostValidationParams) SetValidationID(validationID string) {
	o.ValidationID = validationID
}

// WriteToRequest writes these params to a swagger request
func (o *V2ResetHostValidationParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param host_id
	if err := r.SetPathParam("host_id", o.HostID.String()); err != nil {
		return err
	}

	// path param infra_env_id
	if err := r.SetPathParam("infra_env_id", o.InfraEnvID.String()); err != nil {
		return err
	}

	// path param validation_id
	if err := r.SetPathParam("validation_id", o.ValidationID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}