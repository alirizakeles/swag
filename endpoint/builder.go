// Copyright 2017 Matt Ho
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package endpoint

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/miketonks/swag/swagger"
)

const (
	typeNotSpecified = iota
	typeForm
	typeJSON
)

// Builder uses the builder pattern to generate swagger endpoint definitions
type Builder struct {
	Endpoint  *swagger.Endpoint
	paramType int
}

// ensureParamType ensures we cannot mix form and body data
func (b *Builder) ensureParamType(in string) {
	switch in {
	case "formData":
		if b.paramType == typeJSON {
			panic(fmt.Errorf(`Cannot mix form and body parameters`))
		}
		b.paramType = typeForm
	case "body":
		if b.paramType == typeForm {
			panic(fmt.Errorf(`Cannot mix form and body parameters`))
		}
		b.paramType = typeJSON
	}
}

// Option represents a functional option to customize the swagger endpoint
type Option func(builder *Builder)

// Apply improves the readability of applied options
func (o Option) Apply(builder *Builder) {
	o(builder)
}

// Handler allows an instance of the web handler to be associated with the endpoint.  This can be especially useful when
// using swag to bind the endpoints to the web router.  See the examples package for how the Handler can be used in
// conjunction with Walk to simplify binding endpoints to a router
func Handler(handler interface{}) Option {
	return func(b *Builder) {
		if v, ok := handler.(func(w http.ResponseWriter, r *http.Request)); ok {
			handler = http.HandlerFunc(v)
		}
		b.Endpoint.Handler = handler
	}
}

// Description sets the endpoint's description
func Description(v string) Option {
	return func(b *Builder) {
		b.Endpoint.Description = v
	}
}

// OperationID sets the endpoint's operationId
func OperationID(v string) Option {
	return func(b *Builder) {
		b.Endpoint.OperationID = v
	}
}

// Produces sets the endpoint's produces; by default this will be set to application/json
func Produces(v ...string) Option {
	return func(b *Builder) {
		b.Endpoint.Produces = v
	}
}

// Consumes sets the endpoint's produces; by default this will be set to application/json
func Consumes(v ...string) Option {
	return func(b *Builder) {
		b.Endpoint.Consumes = v
	}
}

func parameter(p swagger.Parameter) Option {
	return func(b *Builder) {
		b.ensureParamType(p.In)
		if b.Endpoint.Parameters == nil {
			b.Endpoint.Parameters = []swagger.Parameter{}
		}

		b.Endpoint.Parameters = append(b.Endpoint.Parameters, p)
	}
}

// Path defines a path parameter for the endpoint; name, typ, format, description, and required correspond to the matching
// swagger fields
func Path(name, typ, format, description string) Option {
	p := swagger.Parameter{
		Name:        name,
		In:          "path",
		Type:        typ,
		Format:      format,
		Description: description,
		Required:    true,
	}
	return parameter(p)
}

// PathMap allows us to define multiple path parameters in a map / struct format
func PathMap(params map[string]swagger.Parameter) Option {
	return func(b *Builder) {
		if b.Endpoint.Parameters == nil {
			b.Endpoint.Parameters = []swagger.Parameter{}
		}

		for k, v := range params {
			v.Name = k
			v.In = "path"
			v.Required = true
			b.Endpoint.Parameters = append(b.Endpoint.Parameters, v)
		}
	}
}

// RequestHeader defines a header parameter for the endpoint; name, typ, format, description, and required correspond to the matching
func RequestHeader(name, typ, format, description string, required bool) Option {
	p := swagger.Parameter{
		Name:        name,
		In:          "header",
		Type:        typ,
		Format:      format,
		Description: description,
		Required:    required,
	}
	return parameter(p)
}

// Query defines a query parameter for the endpoint; name, typ, format, description, and required correspond to the matching
// swagger fields
func Query(name, typ, format, description string, required bool) Option {
	p := swagger.Parameter{
		Name:        name,
		In:          "query",
		Type:        typ,
		Format:      format,
		Description: description,
		Required:    required,
	}
	return parameter(p)
}

// QueryList allows us to define multiple query parameters in a []struct format
func QueryList(params []swagger.Parameter) Option {
	return func(b *Builder) {
		if b.Endpoint.Parameters == nil {
			b.Endpoint.Parameters = []swagger.Parameter{}
		}

		for i, param := range params {
			if param.Name == "" {
				panic(fmt.Errorf(`QueryList parameter %d: %#v has an empty name`, i, param))
			}
			param.In = "query"
			b.Endpoint.Parameters = append(b.Endpoint.Parameters, param)
		}
	}
}

// QueryMap allows us to define multiple query parameters in a map / struct format
func QueryMap(params map[string]swagger.Parameter) Option {
	return func(b *Builder) {
		if b.Endpoint.Parameters == nil {
			b.Endpoint.Parameters = []swagger.Parameter{}
		}

		for k, v := range params {
			v.Name = k
			v.In = "query"
			b.Endpoint.Parameters = append(b.Endpoint.Parameters, v)
		}
	}
}

// FormData defines a form data parameter for the endpoint; name, typ, format, description, and required correspond to the matching
// swagger fields
func FormData(name, typ, format, description string, required bool) Option {
	p := swagger.Parameter{
		Name:        name,
		In:          "formData",
		Type:        typ,
		Format:      format,
		Description: description,
		Required:    required,
	}
	return parameter(p)
}

// FormDataMap allows us to define multiple form data parameters in a map / struct format
func FormDataMap(params map[string]swagger.Parameter) Option {
	return func(b *Builder) {
		b.ensureParamType("formData")
		if b.Endpoint.Parameters == nil {
			b.Endpoint.Parameters = []swagger.Parameter{}
		}

		for k, v := range params {
			v.Name = k
			v.In = "formData"
			b.Endpoint.Parameters = append(b.Endpoint.Parameters, v)
		}
	}
}

// BodyType defines a body parameter for the swagger endpoint as would commonly be used for the POST, PUT, and PATCH methods
// prototype should be a struct or a pointer to struct that swag can use to reflect upon the return type
// t represents the Type of the body
func BodyType(t reflect.Type, description string, required bool) Option {
	p := swagger.Parameter{
		In:          "body",
		Name:        "body",
		Description: description,
		Schema:      swagger.MakeSchema(t),
		Required:    required,
	}
	return parameter(p)
}

// Body defines a body parameter for the swagger endpoint as would commonly be used for the POST, PUT, and PATCH methods
// prototype should be a struct or a pointer to struct that swag can use to reflect upon the return type
func Body(prototype interface{}, description string, required bool) Option {
	return BodyType(reflect.TypeOf(prototype), description, required)
}

// Tags allows one or more tags to be associated with the endpoint
func Tags(tags ...string) Option {
	return func(b *Builder) {
		if b.Endpoint.Tags == nil {
			b.Endpoint.Tags = []string{}
		}

		b.Endpoint.Tags = append(b.Endpoint.Tags, tags...)
	}
}

// Security allows a security scheme to be associated with the endpoint.
func Security(scheme string, scopes ...string) Option {
	return func(b *Builder) {
		if scopes == nil {
			scopes = []string{}
		}
		if b.Endpoint.Security == nil {
			b.Endpoint.Security = &swagger.SecurityRequirement{}
		}

		if b.Endpoint.Security.Requirements == nil {
			b.Endpoint.Security.Requirements = []map[string][]string{}
		}

		b.Endpoint.Security.Requirements = append(b.Endpoint.Security.Requirements, map[string][]string{scheme: scopes})
	}
}

// NoSecurity explicitly sets the endpoint to have no security requirements.
func NoSecurity() Option {
	return func(b *Builder) {
		b.Endpoint.Security = &swagger.SecurityRequirement{DisableSecurity: true}
	}
}

// ResponseOption allows for additional configurations on responses like header information
type ResponseOption func(response *swagger.Response)

// Apply improves the readability of applied options
func (o ResponseOption) Apply(response *swagger.Response) {
	o(response)
}

// Header adds header definitions to swagger responses
func Header(name, typ, format, description string) ResponseOption {
	return func(response *swagger.Response) {
		if response.Headers == nil {
			response.Headers = map[string]swagger.Header{}
		}

		response.Headers[name] = swagger.Header{
			Type:        typ,
			Format:      format,
			Description: description,
		}
	}
}

// HeaderMap allows us to define multiple headers in a map / struct format
func HeaderMap(headers map[string]swagger.Header) ResponseOption {
	return func(response *swagger.Response) {
		if response.Headers == nil {
			response.Headers = map[string]swagger.Header{}
		}

		for k, v := range headers {
			response.Headers[k] = v
		}
	}
}

type fileMarker struct{}

// ResponseFile can be used in an endpoint to change the type to 'file'
var ResponseFile fileMarker

// ResponseType sets the endpoint response for the specified code; may be used multiple times with different status codes
// t represents the Type of the response
func ResponseType(code int, t reflect.Type, description string, opts ...ResponseOption) Option {
	return func(b *Builder) {
		if b.Endpoint.Responses == nil {
			b.Endpoint.Responses = map[string]swagger.Response{}
		}

		r := swagger.Response{
			Description: description,
		}

		if t.Kind() == reflect.Struct && t == reflect.TypeOf(fileMarker{}) {
			r.Schema = &swagger.Schema{
				Type:      "file",
				Prototype: "",
			}
		} else if t.Kind() != reflect.String {
			r.Schema = swagger.MakeSchema(t)
		}

		for _, opt := range opts {
			opt.Apply(&r)
		}

		b.Endpoint.Responses[strconv.Itoa(code)] = r
	}
}

// Response sets the endpoint response for the specified code; may be used multiple times with different status codes
func Response(code int, prototype interface{}, description string, opts ...ResponseOption) Option {
	return ResponseType(code, reflect.TypeOf(prototype), description, opts...)
}

// New constructs a new swagger endpoint using the fields and functional options provided
func New(method, path, summary string, options ...Option) *swagger.Endpoint {
	method = strings.ToUpper(method)
	e := &Builder{
		Endpoint: &swagger.Endpoint{
			Method:      method,
			Path:        path,
			Summary:     summary,
			OperationID: strings.ToLower(method) + camel(path),
			Produces:    []string{"application/json"},
			Consumes:    []string{"application/json"},
			Tags:        []string{},
		},
	}

	for _, opt := range options {
		opt.Apply(e)
	}

	return e.Endpoint
}
