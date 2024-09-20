// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package hi

import (
	"errors"
	"io"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sse"
	"github.com/nbcx/hi/binding"
	"github.com/nbcx/hi/render"
)

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              = binding.MIMEJSON
	MIMEHTML              = binding.MIMEHTML
	MIMEXML               = binding.MIMEXML
	MIMEXML2              = binding.MIMEXML2
	MIMEPlain             = binding.MIMEPlain
	MIMEPOSTForm          = binding.MIMEPOSTForm
	MIMEMultipartPOSTForm = binding.MIMEMultipartPOSTForm
	MIMEYAML              = binding.MIMEYAML
	MIMEYAML2             = binding.MIMEYAML2
	MIMETOML              = binding.MIMETOML
)

// BodyBytesKey indicates a default body bytes key.
const BodyBytesKey = "_gin-gonic/gin/bodybyteskey"

// ContextKey is the key that a Context returns itself for.
const ContextKey = "_gin-gonic/gin/contextkey"

type ContextKeyType int

const ContextRequestKey ContextKeyType = 0

// abortIndex represents a typical value used in abort functions.
const abortIndex int8 = math.MaxInt8 >> 1

type IEngine interface{}

// Context is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type Context struct {
	writermem responseWriter
	Request   *http.Request
	Writer    ResponseWriter

	Params Params
	// handlers HandlersChain[*Context]
	execer Execer // todo: test
	// index    int8
	fullPath string

	// engine       *Engine[IContext] // todo: need check
	// params *Params
	// skippedNodes *[]skippedNode

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]any

	// Errors is a list of errors attached to all the handlers/middlewares who used this context.
	Errors errorMsgs

	// Accepted defines a list of manually accepted formats for content negotiation.
	Accepted []string

	// queryCache caches the query result from c.Request.URL.Query().
	queryCache url.Values

	// formCache caches c.Request.PostForm, which contains the parsed form data from POST, PATCH,
	// or PUT body parameters.
	formCache url.Values

	// SameSite allows a server to define a cookie attribute making it impossible for
	// the browser to send this cookie along with cross-site requests.
	sameSite http.SameSite

	// note: from engin field and func

	// ContextWithFallback enable fallback Context.Deadline(), Context.Done(), Context.Err() and Context.Value() when Context.Request.Context() is not nil.
	ContextWithFallback bool

	// MaxMultipartMemory value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64

	// TrustedPlatform if set to a constant of value gin.Platform*, trusts the headers set by
	// that platform, for example to determine the client IP
	TrustedPlatform string

	// AppEngine was deprecated.
	// Deprecated: USE `TrustedPlatform` WITH VALUE `gin.PlatformGoogleAppEngine` INSTEAD
	// #726 #755 If enabled, it will trust some headers starting with
	// 'X-AppEngine...' for better integration with that PaaS.
	// AppEngine bool

	// RemoteIPHeaders list of headers used to obtain the client IP when
	// `(*gin.Engine).ForwardedByClientIP` is `true` and
	// `(*gin.Context).Request.RemoteAddr` is matched by at least one of the
	// network origins of list defined by `(*gin.Engine).SetTrustedProxies()`.
	RemoteIPHeaders []string

	// ForwardedByClientIP if enabled, client IP will be parsed from the request's headers that
	// match those stored at `(*gin.Engine).RemoteIPHeaders`. If no IP was
	// fetched, it falls back to the IP obtained from
	// `(*gin.Context).Request.RemoteAddr`.
	ForwardedByClientIP bool

	HTMLRender render.HTMLRender
}

func (c *Context) New() {
	// v := make(Params, 0, maxParams)
	// en := e.(*Engine[IContext])
	// skippedNodes := make([]skippedNode, 0, maxSections)

	// c.engine = en
	// c.params = &v
	// c.skippedNodes = &skippedNodes
	c.ForwardedByClientIP = true
	c.RemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}
	c.TrustedPlatform = defaultPlatform
	c.MaxMultipartMemory = defaultMultipartMemory
	// return &Context{engine: engine, params: &v, skippedNodes: &skippedNodes}
}

/************************************/
/********** CONTEXT CREATION ********/
/************************************/
func (c *Context) Init(w http.ResponseWriter, req *http.Request) {
	c.writermem.reset(w)
	c.Request = req
	c.Reset()
}
func (c *Context) Req() *http.Request         { return c.Request }
func (c *Context) WriterMem() *responseWriter { return &c.writermem }
func (c *Context) Rsp() ResponseWriter        { return c.Writer }

// func (c *Context) SetParams(ps *Params)       { c.params = ps }
// func (c *Context) GetParams() *Params         { return c.params }

// func (c *Context) GetSkippedNodes() *[]skippedNode              { return c.skippedNodes }
// func (c *Context) SetHandlers(handlers HandlersChain[*Context]) { c.handlers = handlers }
func (c *Context) SetExecer(execer Execer)     { c.execer = execer }
func (c *Context) GetExecer() Execer           { return c.execer }
func (c *Context) SetFullPath(fullPath string) { c.fullPath = fullPath }

// func (c *Context) GetIndex() int8              { return c.index }
// func (c *Context) SetIndex(index int8)         { c.index = index }
func (c *Context) GetKeys() map[string]any { return c.Keys }
func (c *Context) GetErrors() errorMsgs    { return c.Errors }
func (c *Context) Reset() {
	c.Writer = &c.writermem
	c.Params = c.Params[:0]
	// c.handlers = nil
	c.execer = nil
	// c.index = -1

	c.fullPath = ""
	c.Keys = nil
	c.Errors = c.Errors[:0]
	c.Accepted = nil
	c.queryCache = nil
	c.formCache = nil
	c.sameSite = 0
	// *c.params = (*c.params)[:0]
	// *c.skippedNodes = (*c.skippedNodes)[:0]
}

// Copy returns a copy of the current context that can be safely used outside the request's scope.
// This has to be used when the context has to be passed to a goroutine.
func (c *Context) Copy() *Context {
	cp := Context{
		writermem: c.writermem,
		Request:   c.Request,
		// engine:    c.engine,
	}

	cp.writermem.ResponseWriter = nil
	cp.Writer = &cp.writermem
	// cp.index = abortIndex
	cp.GetExecer().SetIndex(abortIndex)
	cp.execer = nil
	cp.fullPath = c.fullPath

	cKeys := c.Keys
	cp.Keys = make(map[string]any, len(cKeys))
	c.mu.RLock()
	for k, v := range cKeys {
		cp.Keys[k] = v
	}
	c.mu.RUnlock()

	cParams := c.Params
	cp.Params = make([]Param, len(cParams))
	copy(cp.Params, cParams)

	return &cp
}

// HandlerName returns the main handler's name. For example if the handler is "handleGetUsers()",
// this function will return "main.handleGetUsers".
func (c *Context) HandlerName() string {
	// todo: wait check need
	// return nameOfFunction(c.handlers.Last())
	return ""
}

// HandlerNames returns a list of all registered handlers for this context in descending order,
// following the semantics of HandlerName()
func (c *Context) HandlerNames() []string {
	// todo: wait check need
	// hn := make([]string, 0, len(c.handlers))
	// for _, val := range c.handlers {
	// 	if val == nil {
	// 		continue
	// 	}
	// 	hn = append(hn, nameOfFunction(val))
	// }
	// return hn
	return nil
}

// Handler returns the main handler.
func (c *Context) Handler() HandlerFunc[*Context] {
	// todo: wait check need
	// return c.handlers.Last()
	return nil
}

// FullPath returns a matched route full path. For not found routes
// returns an empty string.
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    c.FullPath() == "/user/:id" // true
//	})
func (c *Context) FullPath() string {
	return c.fullPath
}

/************************************/
/*********** FLOW CONTROL ***********/
/************************************/

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in GitHub.
func (c *Context) Next() {
	c.execer.Next()
}

// todo: wait check
type Execer interface {
	Next()
	GetIndex() int8
	SetIndex(int8)
}

func NewExecer[T IContext](ctx T, handlers HandlersChain[T]) Execer {
	return &Exec[T]{ctx: ctx, handlers: handlers, index: -1}
}

type Exec[T IContext] struct {
	index    int8
	handlers HandlersChain[T]
	ctx      T
}

func (c *Exec[T]) GetIndex() int8 {
	return c.index
}

func (c *Exec[T]) SetIndex(index int8) {
	c.index = index
}

func (c *Exec[T]) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		if c.handlers[c.index] == nil {
			continue
		}
		c.handlers[c.index](c.ctx)
		c.index++
	}
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
	return c.execer.GetIndex() >= abortIndex
	// return c.index >= abortIndex
}

// Abort prevents pending handlers from being called. Note that this will not stop the current handler.
// Let's say you have an authorization middleware that validates that the current request is authorized.
// If the authorization fails (ex: the password does not match), call Abort to ensure the remaining handlers
// for this request are not called.
func (c *Context) Abort() {
	// c.index = abortIndex
	c.execer.SetIndex(abortIndex)
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// For example, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Writer.WriteHeaderNow()
	c.Abort()
}

// AbortWithStatusJSON calls `Abort()` and then `JSON` internally.
// This method stops the chain, writes the status code and return a JSON body.
// It also sets the Content-Type as "application/json".
func (c *Context) AbortWithStatusJSON(code int, jsonObj any) {
	c.Abort()
	c.JSON(code, jsonObj)
}

// AbortWithError calls `AbortWithStatus()` and `Error()` internally.
// This method stops the chain, writes the status code and pushes the specified error to `c.Errors`.
// See Context.Error() for more details.
func (c *Context) AbortWithError(code int, err error) *Error {
	c.AbortWithStatus(code)
	return c.Error(err)
}

// Die makes panic of USERSTOPRUN error and go to recover function if defined.
func (c *Context) Die() {
	panic(ErrAbort)
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// For example, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (c *Context) DieWithStatus(code int) {
	c.Status(code)
	c.Writer.WriteHeaderNow()
	c.Die()
}

/************************************/
/********* ERROR MANAGEMENT *********/
/************************************/

// Error attaches an error to the current context. The error is pushed to a list of errors.
// It's a good idea to call Error for each error that occurred during the resolution of a request.
// A middleware can be used to collect all the errors and push them to a database together,
// print a log, or append it in the HTTP response.
// Error will panic if err is nil.
func (c *Context) Error(err error) *Error {
	if err == nil {
		panic("err is nil")
	}

	var parsedError *Error
	ok := errors.As(err, &parsedError)
	if !ok {
		parsedError = &Error{
			Err:  err,
			Type: ErrorTypePrivate,
		}
	}

	c.Errors = append(c.Errors, parsedError)
	return parsedError
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}

	c.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Keys[key]
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString returns the value associated with the key as a string.
func (c *Context) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (c *Context) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (c *Context) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt8 returns the value associated with the key as an integer 8.
func (c *Context) GetInt8(key string) (i8 int8) {
	if val, ok := c.Get(key); ok && val != nil {
		i8, _ = val.(int8)
	}
	return
}

// GetInt16 returns the value associated with the key as an integer 16.
func (c *Context) GetInt16(key string) (i16 int16) {
	if val, ok := c.Get(key); ok && val != nil {
		i16, _ = val.(int16)
	}
	return
}

// GetInt32 returns the value associated with the key as an integer 32.
func (c *Context) GetInt32(key string) (i32 int32) {
	if val, ok := c.Get(key); ok && val != nil {
		i32, _ = val.(int32)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer 64.
func (c *Context) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetUint returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint(key string) (ui uint) {
	if val, ok := c.Get(key); ok && val != nil {
		ui, _ = val.(uint)
	}
	return
}

// GetUint8 returns the value associated with the key as an unsigned integer 8.
func (c *Context) GetUint8(key string) (ui8 uint8) {
	if val, ok := c.Get(key); ok && val != nil {
		ui8, _ = val.(uint8)
	}
	return
}

// GetUint16 returns the value associated with the key as an unsigned integer 16.
func (c *Context) GetUint16(key string) (ui16 uint16) {
	if val, ok := c.Get(key); ok && val != nil {
		ui16, _ = val.(uint16)
	}
	return
}

// GetUint32 returns the value associated with the key as an unsigned integer 32.
func (c *Context) GetUint32(key string) (ui32 uint32) {
	if val, ok := c.Get(key); ok && val != nil {
		ui32, _ = val.(uint32)
	}
	return
}

// GetUint64 returns the value associated with the key as an unsigned integer 64.
func (c *Context) GetUint64(key string) (ui64 uint64) {
	if val, ok := c.Get(key); ok && val != nil {
		ui64, _ = val.(uint64)
	}
	return
}

// GetFloat32 returns the value associated with the key as a float32.
func (c *Context) GetFloat32(key string) (f32 float32) {
	if val, ok := c.Get(key); ok && val != nil {
		f32, _ = val.(float32)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Context) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// GetTime returns the value associated with the key as time.
func (c *Context) GetTime(key string) (t time.Time) {
	if val, ok := c.Get(key); ok && val != nil {
		t, _ = val.(time.Time)
	}
	return
}

// GetDuration returns the value associated with the key as a duration.
func (c *Context) GetDuration(key string) (d time.Duration) {
	if val, ok := c.Get(key); ok && val != nil {
		d, _ = val.(time.Duration)
	}
	return
}

func (c *Context) GetIntSlice(key string) (is []int) {
	if val, ok := c.Get(key); ok && val != nil {
		is, _ = val.([]int)
	}
	return
}

func (c *Context) GetInt8Slice(key string) (i8s []int8) {
	if val, ok := c.Get(key); ok && val != nil {
		i8s, _ = val.([]int8)
	}
	return
}

func (c *Context) GetInt16Slice(key string) (i16s []int16) {
	if val, ok := c.Get(key); ok && val != nil {
		i16s, _ = val.([]int16)
	}
	return
}

func (c *Context) GetInt32Slice(key string) (i32s []int32) {
	if val, ok := c.Get(key); ok && val != nil {
		i32s, _ = val.([]int32)
	}
	return
}

func (c *Context) GetInt64Slice(key string) (i64s []int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64s, _ = val.([]int64)
	}
	return
}

func (c *Context) GetUintSlice(key string) (uis []uint) {
	if val, ok := c.Get(key); ok && val != nil {
		uis, _ = val.([]uint)
	}
	return
}

func (c *Context) GetUint8Slice(key string) (ui8s []uint8) {
	if val, ok := c.Get(key); ok && val != nil {
		ui8s, _ = val.([]uint8)
	}
	return
}

func (c *Context) GetUint16Slice(key string) (ui16s []uint16) {
	if val, ok := c.Get(key); ok && val != nil {
		ui16s, _ = val.([]uint16)
	}
	return
}

func (c *Context) GetUint32Slice(key string) (ui32s []uint32) {
	if val, ok := c.Get(key); ok && val != nil {
		ui32s, _ = val.([]uint32)
	}
	return
}

func (c *Context) GetUint64Slice(key string) (ui64s []uint64) {
	if val, ok := c.Get(key); ok && val != nil {
		ui64s, _ = val.([]uint64)
	}
	return
}

func (c *Context) GetFloat32Slice(key string) (f32s []float32) {
	if val, ok := c.Get(key); ok && val != nil {
		f32s, _ = val.([]float32)
	}
	return
}

func (c *Context) GetFloat64Slice(key string) (f64s []float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64s, _ = val.([]float64)
	}
	return
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (c *Context) GetStringSlice(key string) (ss []string) {
	if val, ok := c.Get(key); ok && val != nil {
		ss, _ = val.([]string)
	}
	return
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (c *Context) GetStringMap(key string) (sm map[string]any) {
	if val, ok := c.Get(key); ok && val != nil {
		sm, _ = val.(map[string]any)
	}
	return
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (c *Context) GetStringMapString(key string) (sms map[string]string) {
	if val, ok := c.Get(key); ok && val != nil {
		sms, _ = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {
	if val, ok := c.Get(key); ok && val != nil {
		smss, _ = val.(map[string][]string)
	}
	return
}

/************************************/
/************ INPUT DATA ************/
/************************************/

// Param returns the value of the URL param.
// It is a shortcut for c.Params.ByName(key)
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    // a GET request to /user/john
//	    id := c.Param("id") // id == "john"
//	    // a GET request to /user/john/
//	    id := c.Param("id") // id == "/john/"
//	})
func (c *Context) Param(key string) string {
	return c.Params.ByName(key)
}

// AddParam adds param to context and
// replaces path param key with given value for e2e testing purposes
// Example Route: "/user/:id"
// AddParam("id", 1)
// Result: "/user/1"
func (c *Context) AddParam(key, value string) {
	c.Params = append(c.Params, Param{Key: key, Value: value})
}

// Query returns the keyed url query value if it exists,
// otherwise it returns an empty string `("")`.
// It is shortcut for `c.Request.URL.Query().Get(key)`
//
//	    GET /path?id=1234&name=Manu&value=
//		   c.Query("id") == "1234"
//		   c.Query("name") == "Manu"
//		   c.Query("value") == ""
//		   c.Query("wtf") == ""
func (c *Context) Query(key string) (value string) {
	value, _ = c.GetQuery(key)
	return
}

// DefaultQuery returns the keyed url query value if it exists,
// otherwise it returns the specified defaultValue string.
// See: Query() and GetQuery() for further information.
//
//	GET /?name=Manu&lastname=
//	c.DefaultQuery("name", "unknown") == "Manu"
//	c.DefaultQuery("id", "none") == "none"
//	c.DefaultQuery("lastname", "none") == ""
func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return defaultValue
}

// GetQuery is like Query(), it returns the keyed url query value
// if it exists `(value, true)` (even when the value is an empty string),
// otherwise it returns `("", false)`.
// It is shortcut for `c.Request.URL.Query().Get(key)`
//
//	GET /?name=Manu&lastname=
//	("Manu", true) == c.GetQuery("name")
//	("", false) == c.GetQuery("id")
//	("", true) == c.GetQuery("lastname")
func (c *Context) GetQuery(key string) (string, bool) {
	if values, ok := c.GetQueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// QueryArray returns a slice of strings for a given query key.
// The length of the slice depends on the number of params with the given key.
func (c *Context) QueryArray(key string) (values []string) {
	values, _ = c.GetQueryArray(key)
	return
}

func (c *Context) initQueryCache() {
	if c.queryCache == nil {
		if c.Request != nil && c.Request.URL != nil {
			c.queryCache = c.Request.URL.Query()
		} else {
			c.queryCache = url.Values{}
		}
	}
}

// GetQueryArray returns a slice of strings for a given query key, plus
// a boolean value whether at least one value exists for the given key.
func (c *Context) GetQueryArray(key string) (values []string, ok bool) {
	c.initQueryCache()
	values, ok = c.queryCache[key]
	return
}

// QueryMap returns a map for a given query key.
func (c *Context) QueryMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetQueryMap(key)
	return
}

// GetQueryMap returns a map for a given query key, plus a boolean value
// whether at least one value exists for the given key.
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

// PostForm returns the specified key from a POST urlencoded form or multipart form
// when it exists, otherwise it returns an empty string `("")`.
func (c *Context) PostForm(key string) (value string) {
	value, _ = c.GetPostForm(key)
	return
}

// DefaultPostForm returns the specified key from a POST urlencoded form or multipart form
// when it exists, otherwise it returns the specified defaultValue string.
// See: PostForm() and GetPostForm() for further information.
func (c *Context) DefaultPostForm(key, defaultValue string) string {
	if value, ok := c.GetPostForm(key); ok {
		return value
	}
	return defaultValue
}

// GetPostForm is like PostForm(key). It returns the specified key from a POST urlencoded
// form or multipart form when it exists `(value, true)` (even when the value is an empty string),
// otherwise it returns ("", false).
// For example, during a PATCH request to update the user's email:
//
//	    email=mail@example.com  -->  ("mail@example.com", true) := GetPostForm("email") // set email to "mail@example.com"
//		   email=                  -->  ("", true) := GetPostForm("email") // set email to ""
//	                            -->  ("", false) := GetPostForm("email") // do nothing with email
func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// PostFormArray returns a slice of strings for a given form key.
// The length of the slice depends on the number of params with the given key.
func (c *Context) PostFormArray(key string) (values []string) {
	values, _ = c.GetPostFormArray(key)
	return
}

func (c *Context) initFormCache() {
	if c.formCache == nil {
		c.formCache = make(url.Values)
		req := c.Request
		if err := req.ParseMultipartForm(c.MaxMultipartMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				debugPrint("error on parse multipart form array: %v", err)
			}
		}
		c.formCache = req.PostForm
	}
}

// GetPostFormArray returns a slice of strings for a given form key, plus
// a boolean value whether at least one value exists for the given key.
func (c *Context) GetPostFormArray(key string) (values []string, ok bool) {
	c.initFormCache()
	values, ok = c.formCache[key]
	return
}

// PostFormMap returns a map for a given form key.
func (c *Context) PostFormMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetPostFormMap(key)
	return
}

// GetPostFormMap returns a map for a given form key, plus a boolean value
// whether at least one value exists for the given key.
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initFormCache()
	return c.get(c.formCache, key)
}

// get is an internal method and returns a map which satisfies conditions.
func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	for k, v := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dicts, exist
}

// FormFile returns the first file for the provided form key.
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(c.MaxMultipartMemory); err != nil {
			return nil, err
		}
	}
	f, fh, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, err
}

// MultipartForm is the parsed multipart form, including file uploads.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(c.MaxMultipartMemory)
	return c.Request.MultipartForm, err
}

// SaveUploadedFile uploads the form file to specific dst.
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// Bind checks the Method and Content-Type to select a binding engine automatically,
// Depending on the "Content-Type" header different bindings are used, for example:
//
//	"application/json" --> JSON binding
//	"application/xml"  --> XML binding
//
// It parses the request's body as JSON if Content-Type == "application/json" using JSON or XML as a JSON input.
// It decodes the json payload into the struct specified as a pointer.
// It writes a 400 error and sets Content-Type header "text/plain" in the response if input is not valid.
func (c *Context) Bind(obj any) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.MustBindWith(obj, b)
}

// BindJSON is a shortcut for c.MustBindWith(obj, binding.JSON).
func (c *Context) BindJSON(obj any) error {
	return c.MustBindWith(obj, binding.JSON)
}

// BindXML is a shortcut for c.MustBindWith(obj, binding.BindXML).
func (c *Context) BindXML(obj any) error {
	return c.MustBindWith(obj, binding.XML)
}

// BindQuery is a shortcut for c.MustBindWith(obj, binding.Query).
func (c *Context) BindQuery(obj any) error {
	return c.MustBindWith(obj, binding.Query)
}

// BindYAML is a shortcut for c.MustBindWith(obj, binding.YAML).
func (c *Context) BindYAML(obj any) error {
	return c.MustBindWith(obj, binding.YAML)
}

// BindTOML is a shortcut for c.MustBindWith(obj, binding.TOML).
func (c *Context) BindTOML(obj any) error {
	return c.MustBindWith(obj, binding.TOML)
}

// BindPlain is a shortcut for c.MustBindWith(obj, binding.Plain).
func (c *Context) BindPlain(obj any) error {
	return c.MustBindWith(obj, binding.Plain)
}

// BindHeader is a shortcut for c.MustBindWith(obj, binding.Header).
func (c *Context) BindHeader(obj any) error {
	return c.MustBindWith(obj, binding.Header)
}

// BindUri binds the passed struct pointer using binding.Uri.
// It will abort the request with HTTP 400 if any error occurs.
func (c *Context) BindUri(obj any) error {
	if err := c.ShouldBindUri(obj); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind) //nolint: errcheck
		return err
	}
	return nil
}

// MustBindWith binds the passed struct pointer using the specified binding engine.
// It will abort the request with HTTP 400 if any error occurs.
// See the binding package.
func (c *Context) MustBindWith(obj any, b binding.Binding) error {
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind) //nolint: errcheck
		return err
	}
	return nil
}

// ShouldBind checks the Method and Content-Type to select a binding engine automatically,
// Depending on the "Content-Type" header different bindings are used, for example:
//
//	"application/json" --> JSON binding
//	"application/xml"  --> XML binding
//
// It parses the request's body as JSON if Content-Type == "application/json" using JSON or XML as a JSON input.
// It decodes the json payload into the struct specified as a pointer.
// Like c.Bind() but this method does not set the response status code to 400 or abort if input is not valid.
func (c *Context) ShouldBind(obj any) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.ShouldBindWith(obj, b)
}

// ShouldBindJSON is a shortcut for c.ShouldBindWith(obj, binding.JSON).
func (c *Context) ShouldBindJSON(obj any) error {
	return c.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindXML is a shortcut for c.ShouldBindWith(obj, binding.XML).
func (c *Context) ShouldBindXML(obj any) error {
	return c.ShouldBindWith(obj, binding.XML)
}

// ShouldBindQuery is a shortcut for c.ShouldBindWith(obj, binding.Query).
func (c *Context) ShouldBindQuery(obj any) error {
	return c.ShouldBindWith(obj, binding.Query)
}

// ShouldBindYAML is a shortcut for c.ShouldBindWith(obj, binding.YAML).
func (c *Context) ShouldBindYAML(obj any) error {
	return c.ShouldBindWith(obj, binding.YAML)
}

// ShouldBindTOML is a shortcut for c.ShouldBindWith(obj, binding.TOML).
func (c *Context) ShouldBindTOML(obj any) error {
	return c.ShouldBindWith(obj, binding.TOML)
}

// ShouldBindPlain is a shortcut for c.ShouldBindWith(obj, binding.Plain).
func (c *Context) ShouldBindPlain(obj any) error {
	return c.ShouldBindWith(obj, binding.Plain)
}

// ShouldBindHeader is a shortcut for c.ShouldBindWith(obj, binding.Header).
func (c *Context) ShouldBindHeader(obj any) error {
	return c.ShouldBindWith(obj, binding.Header)
}

// ShouldBindUri binds the passed struct pointer using the specified binding engine.
func (c *Context) ShouldBindUri(obj any) error {
	m := make(map[string][]string, len(c.Params))
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.Uri.BindUri(m, obj)
}

// ShouldBindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func (c *Context) ShouldBindWith(obj any, b binding.Binding) error {
	return b.Bind(c.Request, obj)
}

// ShouldBindBodyWith is similar with ShouldBindWith, but it stores the request
// body into the context, and reuse when it is called again.
//
// NOTE: This method reads the body before binding. So you should use
// ShouldBindWith for better performance if you need to call only once.
func (c *Context) ShouldBindBodyWith(obj any, bb binding.BindingBody) (err error) {
	var body []byte
	if cb, ok := c.Get(BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			return err
		}
		c.Set(BodyBytesKey, body)
	}
	return bb.BindBody(body, obj)
}

// ShouldBindBodyWithJSON is a shortcut for c.ShouldBindBodyWith(obj, binding.JSON).
func (c *Context) ShouldBindBodyWithJSON(obj any) error {
	return c.ShouldBindBodyWith(obj, binding.JSON)
}

// ShouldBindBodyWithXML is a shortcut for c.ShouldBindBodyWith(obj, binding.XML).
func (c *Context) ShouldBindBodyWithXML(obj any) error {
	return c.ShouldBindBodyWith(obj, binding.XML)
}

// ShouldBindBodyWithYAML is a shortcut for c.ShouldBindBodyWith(obj, binding.YAML).
func (c *Context) ShouldBindBodyWithYAML(obj any) error {
	return c.ShouldBindBodyWith(obj, binding.YAML)
}

// ShouldBindBodyWithTOML is a shortcut for c.ShouldBindBodyWith(obj, binding.TOML).
func (c *Context) ShouldBindBodyWithTOML(obj any) error {
	return c.ShouldBindBodyWith(obj, binding.TOML)
}

// ShouldBindBodyWithJSON is a shortcut for c.ShouldBindBodyWith(obj, binding.JSON).
func (c *Context) ShouldBindBodyWithPlain(obj any) error {
	return c.ShouldBindBodyWith(obj, binding.Plain)
}

// RemoteIP parses the IP from Request.RemoteAddr, normalizes and returns the IP (without the port).
func (c *Context) RemoteIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

// ContentType returns the Content-Type header of the request.
func (c *Context) ContentType() string {
	return filterFlags(c.requestHeader("Content-Type"))
}

// IsWebsocket returns true if the request headers indicate that a websocket
// handshake is being initiated by the client.
func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.requestHeader("Connection")), "upgrade") &&
		strings.EqualFold(c.requestHeader("Upgrade"), "websocket") {
		return true
	}
	return false
}

func (c *Context) requestHeader(key string) string {
	return c.Request.Header.Get(key)
}

/************************************/
/******** RESPONSE RENDERING ********/
/************************************/

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}

// Status sets the HTTP response code.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// Header is an intelligent shortcut for c.Writer.Header().Set(key, value).
// It writes a header in the response.
// If value == "", this method removes the header `c.Writer.Header().Del(key)`
func (c *Context) Header(key, value string) {
	if value == "" {
		c.Writer.Header().Del(key)
		return
	}
	c.Writer.Header().Set(key, value)
}

// GetHeader returns value from request headers.
func (c *Context) GetHeader(key string) string {
	return c.requestHeader(key)
}

// GetRawData returns stream data.
func (c *Context) GetRawData() ([]byte, error) {
	if c.Request.Body == nil {
		return nil, errors.New("cannot read nil body")
	}
	return io.ReadAll(c.Request.Body)
}

// SetSameSite with cookie
func (c *Context) SetSameSite(samesite http.SameSite) {
	c.sameSite = samesite
}

// SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// Cookie returns the named cookie provided in the request or
// ErrNoCookie if not found. And return the named cookie is unescaped.
// If multiple cookies match the given name, only one cookie will
// be returned.
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

// Render writes the response headers and calls render.Render to render data.
func (c *Context) Render(code int, r render.Render) {
	c.Status(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.Writer)
		c.Writer.WriteHeaderNow()
		return
	}

	if err := r.Render(c.Writer); err != nil {
		// Pushing error to c.Errors
		_ = c.Error(err)
		c.Abort()
	}
}

// HTML renders the HTTP template specified by its file name.
// It also updates the HTTP code and sets the Content-Type as "text/html".
// See http://golang.org/doc/articles/wiki/
func (c *Context) HTML(code int, name string, obj any) {
	instance := c.HTMLRender.Instance(name, obj)
	c.Render(code, instance)
}

// IndentedJSON serializes the given struct as pretty JSON (indented + endlines) into the response body.
// It also sets the Content-Type as "application/json".
// WARNING: we recommend using this only for development purposes since printing pretty JSON is
// more CPU and bandwidth consuming. Use Context.JSON() instead.
func (c *Context) IndentedJSON(code int, obj any) {
	c.Render(code, render.IndentedJSON{Data: obj})
}

// SecureJSON serializes the given struct as Secure JSON into the response body.
// Default prepends "while(1)," to response body if the given struct is array values.
// It also sets the Content-Type as "application/json".
func (c *Context) SecureJSON(code int, obj any, secureJSONPrefix ...string) {
	prefix := "while(1);"
	if len(secureJSONPrefix) > 0 {
		prefix = secureJSONPrefix[0]
	}
	c.Render(code, render.SecureJSON{Prefix: prefix, Data: obj})
}

// JSONP serializes the given struct as JSON into the response body.
// It adds padding to response body to request data from a server residing in a different domain than the client.
// It also sets the Content-Type as "application/javascript".
func (c *Context) JSONP(code int, obj any) {
	callback := c.DefaultQuery("callback", "")
	if callback == "" {
		c.Render(code, render.JSON{Data: obj})
		return
	}
	c.Render(code, render.JsonpJSON{Callback: callback, Data: obj})
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj any) {
	c.Render(code, render.JSON{Data: obj})
}

// AsciiJSON serializes the given struct as JSON into the response body with unicode to ASCII string.
// It also sets the Content-Type as "application/json".
func (c *Context) AsciiJSON(code int, obj any) {
	c.Render(code, render.AsciiJSON{Data: obj})
}

// PureJSON serializes the given struct as JSON into the response body.
// PureJSON, unlike JSON, does not replace special html characters with their unicode entities.
func (c *Context) PureJSON(code int, obj any) {
	c.Render(code, render.PureJSON{Data: obj})
}

// XML serializes the given struct as XML into the response body.
// It also sets the Content-Type as "application/xml".
func (c *Context) XML(code int, obj any) {
	c.Render(code, render.XML{Data: obj})
}

// YAML serializes the given struct as YAML into the response body.
func (c *Context) YAML(code int, obj any) {
	c.Render(code, render.YAML{Data: obj})
}

// TOML serializes the given struct as TOML into the response body.
func (c *Context) TOML(code int, obj any) {
	c.Render(code, render.TOML{Data: obj})
}

// ProtoBuf serializes the given struct as ProtoBuf into the response body.
func (c *Context) ProtoBuf(code int, obj any) {
	c.Render(code, render.ProtoBuf{Data: obj})
}

// String writes the given string into the response body.
func (c *Context) String(code int, format string, values ...any) {
	c.Render(code, render.String{Format: format, Data: values})
}

// Redirect returns an HTTP redirect to the specific location.
func (c *Context) Redirect(code int, location string) {
	c.Render(-1, render.Redirect{
		Code:     code,
		Location: location,
		Request:  c.Request,
	})
}

// Data writes some data into the body stream and updates the HTTP code.
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, render.Data{
		ContentType: contentType,
		Data:        data,
	})
}

// DataFromReader writes the specified reader into the body stream and updates the HTTP code.
func (c *Context) DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string) {
	c.Render(code, render.Reader{
		Headers:       extraHeaders,
		ContentType:   contentType,
		ContentLength: contentLength,
		Reader:        reader,
	})
}

// File writes the specified file into the body stream in an efficient way.
func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// FileFromFS writes the specified file from http.FileSystem into the body stream in an efficient way.
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.Request.URL.Path = old
	}(c.Request.URL.Path)

	c.Request.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.Writer, c.Request)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// FileAttachment writes the specified file into the body stream in an efficient way
// On the client side, the file will typically be downloaded with the given filename
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+escapeQuotes(filename)+`"`)
	} else {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.Writer, c.Request, filepath)
}

// SSEvent writes a Server-Sent Event into the body stream.
func (c *Context) SSEvent(name string, message any) {
	c.Render(-1, sse.Event{
		Event: name,
		Data:  message,
	})
}

// Stream sends a streaming response and returns a boolean
// indicates "Is client disconnected in middle of stream"
func (c *Context) Stream(step func(w io.Writer) bool) bool {
	w := c.Writer
	clientGone := w.CloseNotify()
	for {
		select {
		case <-clientGone:
			return true
		default:
			keepOpen := step(w)
			w.Flush()
			if !keepOpen {
				return false
			}
		}
	}
}

/************************************/
/******** CONTENT NEGOTIATION *******/
/************************************/

// Negotiate contains all negotiations data.
type Negotiate struct {
	Offered  []string
	HTMLName string
	HTMLData any
	JSONData any
	XMLData  any
	YAMLData any
	Data     any
	TOMLData any
}

// Negotiate calls different Render according to acceptable Accept format.
func (c *Context) Negotiate(code int, config Negotiate) {
	switch c.NegotiateFormat(config.Offered...) {
	case binding.MIMEJSON:
		data := chooseData(config.JSONData, config.Data)
		c.JSON(code, data)

	case binding.MIMEHTML:
		data := chooseData(config.HTMLData, config.Data)
		c.HTML(code, config.HTMLName, data)

	case binding.MIMEXML:
		data := chooseData(config.XMLData, config.Data)
		c.XML(code, data)

	case binding.MIMEYAML, binding.MIMEYAML2:
		data := chooseData(config.YAMLData, config.Data)
		c.YAML(code, data)

	case binding.MIMETOML:
		data := chooseData(config.TOMLData, config.Data)
		c.TOML(code, data)

	default:
		c.AbortWithError(http.StatusNotAcceptable, errors.New("the accepted formats are not offered by the server")) //nolint: errcheck
	}
}

// NegotiateFormat returns an acceptable Accept format.
func (c *Context) NegotiateFormat(offered ...string) string {
	assert1(len(offered) > 0, "you must provide at least one offer")

	if c.Accepted == nil {
		c.Accepted = parseAccept(c.requestHeader("Accept"))
	}
	if len(c.Accepted) == 0 {
		return offered[0]
	}
	for _, accepted := range c.Accepted {
		for _, offer := range offered {
			// According to RFC 2616 and RFC 2396, non-ASCII characters are not allowed in headers,
			// therefore we can just iterate over the string without casting it into []rune
			i := 0
			for ; i < len(accepted) && i < len(offer); i++ {
				if accepted[i] == '*' || offer[i] == '*' {
					return offer
				}
				if accepted[i] != offer[i] {
					break
				}
			}
			if i == len(accepted) {
				return offer
			}
		}
	}
	return ""
}

// SetAccepted sets Accept header data.
func (c *Context) SetAccepted(formats ...string) {
	c.Accepted = formats
}

/************************************/
/***** GOLANG.ORG/X/NET/CONTEXT *****/
/************************************/

// hasRequestContext returns whether c.Request has Context and fallback.
func (c *Context) hasRequestContext() bool {
	hasFallback := c.ContextWithFallback
	hasRequestContext := c.Request != nil && c.Request.Context() != nil
	return hasFallback && hasRequestContext
}

// Deadline returns that there is no deadline (ok==false) when c.Request has no Context.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	if !c.hasRequestContext() {
		return
	}
	return c.Request.Context().Deadline()
}

// Done returns nil (chan which will wait forever) when c.Request has no Context.
func (c *Context) Done() <-chan struct{} {
	if !c.hasRequestContext() {
		return nil
	}
	return c.Request.Context().Done()
}

// Err returns nil when c.Request has no Context.
func (c *Context) Err() error {
	if !c.hasRequestContext() {
		return nil
	}
	return c.Request.Context().Err()
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *Context) Value(key any) any {
	if key == ContextRequestKey {
		return c.Request
	}
	if key == ContextKey {
		return c
	}
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	if !c.hasRequestContext() {
		return nil
	}
	return c.Request.Context().Value(key)
}

//****** todo: 此函数实现过于复杂，后续需优化

// ClientIP implements one best effort algorithm to return the real client IP.
// It calls c.RemoteIP() under the hood, to check if the remote IP is a trusted proxy or not.
// If it is it will then try to parse the headers defined in Engine.RemoteIPHeaders (defaulting to [X-Forwarded-For, X-Real-Ip]).
// If the headers are not syntactically valid OR the remote IP does not correspond to a trusted proxy,
// the remote IP (coming from Request.RemoteAddr) is returned.
func (c *Context) ClientIP() string {
	// Check if we're running on a trusted platform, continue running backwards if error
	if c.TrustedPlatform != "" {
		// Developers can define their own header of Trusted Platform or use predefined constants
		if addr := c.requestHeader(c.TrustedPlatform); addr != "" {
			return addr
		}
	}

	// Legacy "AppEngine" flag
	// if c.AppEngine {
	// 	log.Println(`The AppEngine flag is going to be deprecated. Please check issues #2723 and #2739 and use 'TrustedPlatform: gin.PlatformGoogleAppEngine' instead.`)
	// 	if addr := c.requestHeader("X-Appengine-Remote-Addr"); addr != "" {
	// 		return addr
	// 	}
	// }

	// It also checks if the remoteIP is a trusted proxy or not.
	// In order to perform this validation, it will see if the IP is contained within at least one of the CIDR blocks
	// defined by Engine.SetTrustedProxies()
	remoteIP := net.ParseIP(c.RemoteIP())
	if remoteIP == nil {
		return ""
	}
	trusted := c.isTrustedProxy(remoteIP)

	if trusted && c.ForwardedByClientIP && c.RemoteIPHeaders != nil {
		for _, headerName := range c.RemoteIPHeaders {
			ip, valid := c.validateHeader(c.requestHeader(headerName))
			if valid {
				return ip
			}
		}
	}
	return remoteIP.String()
}

// isTrustedProxy will check whether the IP address is included in the trusted list according to Engine.trustedCIDRs
func (c *Context) isTrustedProxy(ip net.IP) bool {
	// 	if engine.trustedCIDRs == nil {
	// 		return false
	// 	}
	if defaultTrustedCIDRs == nil {
		return false
	}
	for _, cidr := range defaultTrustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// validateHeader will parse X-Forwarded-For header and return the trusted client IP address
func (c *Context) validateHeader(header string) (clientIP string, valid bool) {
	if header == "" {
		return "", false
	}
	items := strings.Split(header, ",")
	for i := len(items) - 1; i >= 0; i-- {
		ipStr := strings.TrimSpace(items[i])
		ip := net.ParseIP(ipStr)
		if ip == nil {
			break
		}

		// X-Forwarded-For is appended by proxy
		// Check IPs in reverse order and stop when find untrusted proxy
		if (i == 0) || (!c.isTrustedProxy(ip)) {
			return ipStr, true
		}
	}
	return "", false
}

// func (engine *Engine[T]) prepareTrustedCIDRs() ([]*net.IPNet, error) {
// 	if engine.trustedProxies == nil {
// 		return nil, nil
// 	}

// 	cidr := make([]*net.IPNet, 0, len(engine.trustedProxies))
// 	for _, trustedProxy := range engine.trustedProxies {
// 		if !strings.Contains(trustedProxy, "/") {
// 			ip := parseIP(trustedProxy)
// 			if ip == nil {
// 				return cidr, &net.ParseError{Type: "IP address", Text: trustedProxy}
// 			}

// 			switch len(ip) {
// 			case net.IPv4len:
// 				trustedProxy += "/32"
// 			case net.IPv6len:
// 				trustedProxy += "/128"
// 			}
// 		}
// 		_, cidrNet, err := net.ParseCIDR(trustedProxy)
// 		if err != nil {
// 			return cidr, err
// 		}
// 		cidr = append(cidr, cidrNet)
// 	}
// 	return cidr, nil
// }

// SetTrustedProxies set a list of network origins (IPv4 addresses,
// IPv4 CIDRs, IPv6 addresses or IPv6 CIDRs) from which to trust
// request's headers that contain alternative client IP when
// `(*gin.Engine).ForwardedByClientIP` is `true`. `TrustedProxies`
// feature is enabled by default, and it also trusts all proxies
// by default. If you want to disable this feature, use
// Engine.SetTrustedProxies(nil), then Context.ClientIP() will
// return the remote address directly.
// func (engine *Engine[T]) SetTrustedProxies(trustedProxies []string) error {
// 	engine.trustedProxies = trustedProxies
// 	return engine.parseTrustedProxies()
// }

// isUnsafeTrustedProxies checks if Engine.trustedCIDRs contains all IPs, it's not safe if it has (returns true)
// func (engine *Engine[T]) isUnsafeTrustedProxies() bool {
// 	return engine.isTrustedProxy(net.ParseIP("0.0.0.0")) || engine.isTrustedProxy(net.ParseIP("::"))
// }

// parseTrustedProxies parse Engine.trustedProxies to Engine.trustedCIDRs
// func (engine *Engine[T]) parseTrustedProxies() error {
// 	trustedCIDRs, err := engine.prepareTrustedCIDRs()
// 	engine.trustedCIDRs = trustedCIDRs
// 	return err
// }
