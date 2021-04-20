// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package tlog

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sigstore/rekor/pkg/generated/models"
)

// GetPublicKeyOKCode is the HTTP code returned for type GetPublicKeyOK
const GetPublicKeyOKCode int = 200

/*GetPublicKeyOK The public key

swagger:response getPublicKeyOK
*/
type GetPublicKeyOK struct {

	/*
	  In: Body
	*/
	Payload string `json:"body,omitempty"`
}

// NewGetPublicKeyOK creates GetPublicKeyOK with default headers values
func NewGetPublicKeyOK() *GetPublicKeyOK {

	return &GetPublicKeyOK{}
}

// WithPayload adds the payload to the get public key o k response
func (o *GetPublicKeyOK) WithPayload(payload string) *GetPublicKeyOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get public key o k response
func (o *GetPublicKeyOK) SetPayload(payload string) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPublicKeyOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

/*GetPublicKeyDefault There was an internal error in the server while processing the request

swagger:response getPublicKeyDefault
*/
type GetPublicKeyDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetPublicKeyDefault creates GetPublicKeyDefault with default headers values
func NewGetPublicKeyDefault(code int) *GetPublicKeyDefault {
	if code <= 0 {
		code = 500
	}

	return &GetPublicKeyDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get public key default response
func (o *GetPublicKeyDefault) WithStatusCode(code int) *GetPublicKeyDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get public key default response
func (o *GetPublicKeyDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get public key default response
func (o *GetPublicKeyDefault) WithPayload(payload *models.Error) *GetPublicKeyDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get public key default response
func (o *GetPublicKeyDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPublicKeyDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
