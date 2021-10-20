// Package Openapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.8.2 DO NOT EDIT.
package Openapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// Configuration for a container that is portable between hosts
type ContainerConfig struct {
	Cmd *[]string `json:"cmd,omitempty"`

	// List of environment variables set when the command is executed
	Env *[]string `json:"env,omitempty"`

	// The name of the image to use when creating the container
	Image *string `json:"image,omitempty"`

	// List of jail parameters (see jail(8) for details)
	JailParam *[]string `json:"jail_param,omitempty"`

	// List of networks that the container should be connected to
	Networks *[]string `json:"networks,omitempty"`

	// List of volumes that should be mounted into the container
	Volumes *[]string `json:"volumes,omitempty"`
}

// summary description of a container
type ContainerSummary struct {
	// Command being used when starting the container
	Command *string `json:"command,omitempty"`

	// When the container was created
	Created *string `json:"created,omitempty"`

	// The id of this container
	Id *string `json:"id,omitempty"`

	// The id of the image that this container was created from
	ImageId *string `json:"image_id,omitempty"`

	// Name of the image that this container was created from
	ImageName *string `json:"image_name,omitempty"`

	// Tag of the image that this container was created from
	ImageTag *string `json:"image_tag,omitempty"`

	// Name of the container
	Name *string `json:"name,omitempty"`

	// whether or not the container is running
	Running *bool `json:"running,omitempty"`
}

// Represents an error
type ErrorResponse struct {
	// The error message.
	Message string `json:"message"`
}

// Response to an API call that returns just an Id
type IdResponse struct {
	// The id of the created/modified/destroyed object.
	Id string `json:"id"`
}

// ContainerCreateJSONBody defines parameters for ContainerCreate.
type ContainerCreateJSONBody ContainerConfig

// ContainerCreateParams defines parameters for ContainerCreate.
type ContainerCreateParams struct {
	// Assign the specified name to the container. Must match `/?[a-zA-Z0-9][a-zA-Z0-9_.-]+`.
	Name *string `json:"name,omitempty"`
}

// ContainerListParams defines parameters for ContainerList.
type ContainerListParams struct {
	// Return all containers. By default, only running containers are shown.
	All *bool `json:"all,omitempty"`
}

// ContainerCreateJSONRequestBody defines body for ContainerCreate for application/json ContentType.
type ContainerCreateJSONRequestBody ContainerCreateJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// ContainerCreate request with any body
	ContainerCreateWithBody(ctx context.Context, params *ContainerCreateParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ContainerCreate(ctx context.Context, params *ContainerCreateParams, body ContainerCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ContainerList request
	ContainerList(ctx context.Context, params *ContainerListParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ContainerDelete request
	ContainerDelete(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ContainerStart request
	ContainerStart(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ContainerStop request
	ContainerStop(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) ContainerCreateWithBody(ctx context.Context, params *ContainerCreateParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewContainerCreateRequestWithBody(c.Server, params, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ContainerCreate(ctx context.Context, params *ContainerCreateParams, body ContainerCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewContainerCreateRequest(c.Server, params, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ContainerList(ctx context.Context, params *ContainerListParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewContainerListRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ContainerDelete(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewContainerDeleteRequest(c.Server, containerId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ContainerStart(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewContainerStartRequest(c.Server, containerId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ContainerStop(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewContainerStopRequest(c.Server, containerId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewContainerCreateRequest calls the generic ContainerCreate builder with application/json body
func NewContainerCreateRequest(server string, params *ContainerCreateParams, body ContainerCreateJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewContainerCreateRequestWithBody(server, params, "application/json", bodyReader)
}

// NewContainerCreateRequestWithBody generates requests for ContainerCreate with any type of body
func NewContainerCreateRequestWithBody(server string, params *ContainerCreateParams, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/containers/create")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.Name != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "name", runtime.ParamLocationQuery, *params.Name); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewContainerListRequest generates requests for ContainerList
func NewContainerListRequest(server string, params *ContainerListParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/containers/list")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.All != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "all", runtime.ParamLocationQuery, *params.All); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewContainerDeleteRequest generates requests for ContainerDelete
func NewContainerDeleteRequest(server string, containerId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "container_id", runtime.ParamLocationPath, containerId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/containers/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewContainerStartRequest generates requests for ContainerStart
func NewContainerStartRequest(server string, containerId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "container_id", runtime.ParamLocationPath, containerId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/containers/%s/start", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewContainerStopRequest generates requests for ContainerStop
func NewContainerStopRequest(server string, containerId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "container_id", runtime.ParamLocationPath, containerId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/containers/%s/stop", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// ContainerCreate request with any body
	ContainerCreateWithBodyWithResponse(ctx context.Context, params *ContainerCreateParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ContainerCreateResponse, error)

	ContainerCreateWithResponse(ctx context.Context, params *ContainerCreateParams, body ContainerCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*ContainerCreateResponse, error)

	// ContainerList request
	ContainerListWithResponse(ctx context.Context, params *ContainerListParams, reqEditors ...RequestEditorFn) (*ContainerListResponse, error)

	// ContainerDelete request
	ContainerDeleteWithResponse(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*ContainerDeleteResponse, error)

	// ContainerStart request
	ContainerStartWithResponse(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*ContainerStartResponse, error)

	// ContainerStop request
	ContainerStopWithResponse(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*ContainerStopResponse, error)
}

type ContainerCreateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON201      *IdResponse
}

// Status returns HTTPResponse.Status
func (r ContainerCreateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ContainerCreateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ContainerListResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]ContainerSummary
}

// Status returns HTTPResponse.Status
func (r ContainerListResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ContainerListResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ContainerDeleteResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *IdResponse
	JSON404      *ErrorResponse
	JSON500      *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r ContainerDeleteResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ContainerDeleteResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ContainerStartResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *IdResponse
	JSON304      *ErrorResponse
	JSON404      *ErrorResponse
	JSON500      *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r ContainerStartResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ContainerStartResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ContainerStopResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *IdResponse
	JSON304      *ErrorResponse
	JSON404      *ErrorResponse
	JSON500      *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r ContainerStopResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ContainerStopResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ContainerCreateWithBodyWithResponse request with arbitrary body returning *ContainerCreateResponse
func (c *ClientWithResponses) ContainerCreateWithBodyWithResponse(ctx context.Context, params *ContainerCreateParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ContainerCreateResponse, error) {
	rsp, err := c.ContainerCreateWithBody(ctx, params, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseContainerCreateResponse(rsp)
}

func (c *ClientWithResponses) ContainerCreateWithResponse(ctx context.Context, params *ContainerCreateParams, body ContainerCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*ContainerCreateResponse, error) {
	rsp, err := c.ContainerCreate(ctx, params, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseContainerCreateResponse(rsp)
}

// ContainerListWithResponse request returning *ContainerListResponse
func (c *ClientWithResponses) ContainerListWithResponse(ctx context.Context, params *ContainerListParams, reqEditors ...RequestEditorFn) (*ContainerListResponse, error) {
	rsp, err := c.ContainerList(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseContainerListResponse(rsp)
}

// ContainerDeleteWithResponse request returning *ContainerDeleteResponse
func (c *ClientWithResponses) ContainerDeleteWithResponse(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*ContainerDeleteResponse, error) {
	rsp, err := c.ContainerDelete(ctx, containerId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseContainerDeleteResponse(rsp)
}

// ContainerStartWithResponse request returning *ContainerStartResponse
func (c *ClientWithResponses) ContainerStartWithResponse(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*ContainerStartResponse, error) {
	rsp, err := c.ContainerStart(ctx, containerId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseContainerStartResponse(rsp)
}

// ContainerStopWithResponse request returning *ContainerStopResponse
func (c *ClientWithResponses) ContainerStopWithResponse(ctx context.Context, containerId string, reqEditors ...RequestEditorFn) (*ContainerStopResponse, error) {
	rsp, err := c.ContainerStop(ctx, containerId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseContainerStopResponse(rsp)
}

// ParseContainerCreateResponse parses an HTTP response from a ContainerCreateWithResponse call
func ParseContainerCreateResponse(rsp *http.Response) (*ContainerCreateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ContainerCreateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest IdResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	}

	return response, nil
}

// ParseContainerListResponse parses an HTTP response from a ContainerListWithResponse call
func ParseContainerListResponse(rsp *http.Response) (*ContainerListResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ContainerListResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []ContainerSummary
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseContainerDeleteResponse parses an HTTP response from a ContainerDeleteWithResponse call
func ParseContainerDeleteResponse(rsp *http.Response) (*ContainerDeleteResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ContainerDeleteResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest IdResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseContainerStartResponse parses an HTTP response from a ContainerStartWithResponse call
func ParseContainerStartResponse(rsp *http.Response) (*ContainerStartResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ContainerStartResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest IdResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 304:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON304 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseContainerStopResponse parses an HTTP response from a ContainerStopWithResponse call
func ParseContainerStopResponse(rsp *http.Response) (*ContainerStopResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ContainerStopResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest IdResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 304:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON304 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Create a container
	// (POST /containers/create)
	ContainerCreate(ctx echo.Context, params ContainerCreateParams) error
	// List containers
	// (GET /containers/list)
	ContainerList(ctx echo.Context, params ContainerListParams) error
	// Remove a container
	// (DELETE /containers/{container_id})
	ContainerDelete(ctx echo.Context, containerId string) error
	// Start a container
	// (POST /containers/{container_id}/start)
	ContainerStart(ctx echo.Context, containerId string) error
	// Stop a container
	// (POST /containers/{container_id}/stop)
	ContainerStop(ctx echo.Context, containerId string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// ContainerCreate converts echo context to params.
func (w *ServerInterfaceWrapper) ContainerCreate(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params ContainerCreateParams
	// ------------- Optional query parameter "name" -------------

	err = runtime.BindQueryParameter("form", true, false, "name", ctx.QueryParams(), &params.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ContainerCreate(ctx, params)
	return err
}

// ContainerList converts echo context to params.
func (w *ServerInterfaceWrapper) ContainerList(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params ContainerListParams
	// ------------- Optional query parameter "all" -------------

	err = runtime.BindQueryParameter("form", true, false, "all", ctx.QueryParams(), &params.All)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter all: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ContainerList(ctx, params)
	return err
}

// ContainerDelete converts echo context to params.
func (w *ServerInterfaceWrapper) ContainerDelete(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "container_id" -------------
	var containerId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "container_id", runtime.ParamLocationPath, ctx.Param("container_id"), &containerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter container_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ContainerDelete(ctx, containerId)
	return err
}

// ContainerStart converts echo context to params.
func (w *ServerInterfaceWrapper) ContainerStart(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "container_id" -------------
	var containerId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "container_id", runtime.ParamLocationPath, ctx.Param("container_id"), &containerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter container_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ContainerStart(ctx, containerId)
	return err
}

// ContainerStop converts echo context to params.
func (w *ServerInterfaceWrapper) ContainerStop(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "container_id" -------------
	var containerId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "container_id", runtime.ParamLocationPath, ctx.Param("container_id"), &containerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter container_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ContainerStop(ctx, containerId)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.POST(baseURL+"/containers/create", wrapper.ContainerCreate)
	router.GET(baseURL+"/containers/list", wrapper.ContainerList)
	router.DELETE(baseURL+"/containers/:container_id", wrapper.ContainerDelete)
	router.POST(baseURL+"/containers/:container_id/start", wrapper.ContainerStart)
	router.POST(baseURL+"/containers/:container_id/stop", wrapper.ContainerStop)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xX72/bNhP+Vw583w8tXll237RbZyAY0ibrsnVFkaQYsCBIafFsMZVIlaTseIH/9+FI",
	"25IsJXYQY1/aT7ap8/147p7jozuW6LzQCpWzbHjHbJJizv3Xt1o5LhWat1qN5YSOBNrEyMJJrdiQhfPS",
	"cPoNY22AQ7L6E7iUO5AWCm0cH2UII3QzRAWpts6yiBVGF2icRB8tyQV94C3PiwzZ8JL1R1L1bcoi1ktY",
	"xDILfXYVMekw9/9w8wLZkFlnpJqwRbQ64MbwOf1GNW0n/V5aB3oMqKbSaJWjcjDlRlKKFiw6mKWowKUI",
	"ic5zrgQVgbeYlA4Fi+oZHp+8+fTucMAi9v7ow7tDwa+Pf48/XfzSe/24RGXOJ9hO9SJFUDxHSpfy8Wbg",
	"NJQWQ5aJQe6kmizTXULPonbIGy6z64Ibnt8PCdmAt0GHxsIzi+gPn71+7tsr0HGZ2edNFHiW6Vls+Oza",
	"6uQLOnvoTIksYtoazJBbPLyhB4ZcPQ4YhW6mzRd7f84rizBuDRjAprrMBIz8mcLEoQCn2WMSmOqszPGB",
	"+EuDEL4KmOtSUTipnG41Z9fwdCAdgdzi4tpUj24wcSxitz3rTJk4NmQnmbyVJv7Ngx6fqIlUGB99PI3P",
	"A7njTW+LqApwXuY5N/N2xTY8gNopAcAbpW1wOhCoa3MEZo2Qhre0KMI8W8fNbvPsJx87XP9ZsXc1BzNu",
	"YWXf4UqKbupJEYgn7cOpeFpeb/Oypm+Y07rTeoIwNjq/Pwqtg3acD+0l8aQojnds+ws+2U+M7TU8CLcp",
	"laKvLQ+zFF2KBrQBpTd3gbSw+uPa50jrDLnqJtqKB3ti2srdImInxmhzhrbQynYAcYaFQUv3MXAFSMb1",
	"fXvHcrTW3xfsXOfoUuLLjK6xmdFqElOIJg3Xf+iaTx8AliZxG3BCHL+W0hDZLte+rirImvU8Da+mr0XE",
	"TsVDSIUndCVyBUcfTyHhWRYG06ArjbJwU1pHT09Faz1tp+xylPu5FnIsUfQFWmf0HAWE8rYDJkUdq1o5",
	"TwOq5mhBIaUa63Y1v15cfPS40O1N27hUMgmKYSZd6msMMSDEYFWqjXM4Ozm/IE9U8BSNDf4H8SB+QW3S",
	"BSpeSDZkB/EgJkVUcJd6lPtrEtp+wJNOC20dfVI/vH48FY1rLhiSm5UcYcPLzfKOrJWTsO1tgYlvUVBM",
	"m7duDH/QHOTcJSl87v98yXt/H/X+GvR+uqq+Xse9q/99pgolef9aoud/2FfhI1oK5I7be3EVGo/WvdFi",
	"Hu4/5VD5OnlRZB56rfo3lrK/q7n6r8ExG7L/9Csx3l8q8X7rsqZ2t4T4cs0lDUm+k1SshpVkm5/eMFi+",
	"e/8fvNhbJZsz2yxC6eWyo0d2pUJYmISGzCCD+lRlMszSBF3Xkgh7gBzkBU8ckDkBoceVT0tdv2cWSept",
	"m8QQBWj71HzCG5JLY15mLgKtsvnqDqoZATdIwnGm7hs8nmVdc1fdXletng0e1bO1Ht1pDGs32YZc3bWl",
	"XjtXCLT6ebf+fi3FIrQ0w7A37unRcTDY0qXTYy8OupRGDEcKpJJO8gwsTvx74UruCEi4IllvSwKSdP0Y",
	"pINSya8lZtRkhyaXyr8I1L2uWkrbsOpovb4WA7eumCd0eh/sjNjLwcsdonYqlg8abJmkFUJDEOMfULx8",
	"dXDw6scRed8t2w2h0JlwMxL5frVHvLZmYNFM0XQy4AxzPX14qTVJ0PdvRjtcnOfe7jsP/g0eHOzEgz1N",
	"U/UqwzODXMzDyzKK74zcByM9bx5HSF3sxEddfKfjt0FHXRRPpmMnSb4NAupig3+LxT8BAAD//w6VOS+e",
	"GAAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}

