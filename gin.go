// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package hi

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/nbcx/hi/internal/bytesconv"
	"github.com/nbcx/hi/render"

	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type IContext interface {
	New(maxSections, maxParams uint16)
	Init(w http.ResponseWriter, req *http.Request)
	Req() *http.Request
	Rsp() ResponseWriter
	Next()
	WriterMem() *responseWriter
	SetParams(*Params)
	GetParams() *Params
	// GetSkippedNodes() *[]SkippedNode
	// SetHandlers(HandlersChain)
	SetExecer(Execer)
	GetExecer() Execer
	// GetIndex() int8
	// SetIndex(int8)

	SetFullPath(string)
	Reset()
	Bind(obj any) error
	Header(key, value string)
	AbortWithStatus(code int)
	Set(key string, value any)
	GetKeys() map[string]any
	GetErrors() errorMsgs
	ClientIP() string
	JSON(code int, obj any)
	Abort()
	Error(err error) *Error
	Param(key string) string
	FileFromFS(filepath string, fs http.FileSystem)
	File(filepath string)
}

const defaultMultipartMemory = 32 << 20 // 32 MB
const escapedColon = "\\:"
const colon = ":"
const backslash = "\\"

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

var defaultPlatform string

var defaultTrustedCIDRs = []*net.IPNet{
	{ // 0.0.0.0/0 (IPv4)
		IP:   net.IP{0x0, 0x0, 0x0, 0x0},
		Mask: net.IPMask{0x0, 0x0, 0x0, 0x0},
	},
	{ // ::/0 (IPv6)
		IP:   net.IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Mask: net.IPMask{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	},
}

var regSafePrefix = regexp.MustCompile("[^a-zA-Z0-9/-]+")
var regRemoveRepeatedChar = regexp.MustCompile("/{2,}")

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc[T IContext] func(T)

// OptionFunc defines the function to change the default configuration
type OptionFunc[T IContext] func(*Engine[T])

// HandlersChain defines a HandlerFunc slice.
type HandlersChain[T IContext] []HandlerFunc[T]

// Last returns the last handler in the chain. i.e. the last handler is the main one.
func (c HandlersChain[T]) Last() HandlerFunc[T] {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// RouteInfo represents a request route's specification which contains method and path and its handler.
type RouteInfo[T IContext] struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc HandlerFunc[T]
}

// RoutesInfo defines a RouteInfo slice.
type RoutesInfo[T IContext] []RouteInfo[T]

// Trusted platforms
const (
	// PlatformGoogleAppEngine when running on Google App Engine. Trust X-Appengine-Remote-Addr
	// for determining the client's IP
	PlatformGoogleAppEngine = "X-Appengine-Remote-Addr"
	// PlatformCloudflare when using Cloudflare's CDN. Trust CF-Connecting-IP for determining
	// the client's IP
	PlatformCloudflare = "CF-Connecting-IP"
	// PlatformFlyIO when running on Fly.io. Trust Fly-Client-IP for determining the client's IP
	PlatformFlyIO = "Fly-Client-IP"
)

// Engine is the framework's instance, it contains the muxer, middleware and configuration settings.
// Create an instance of Engine, by using New() or Default()
type Engine[T IContext] struct {
	RouterGroup[T]

	// RedirectTrailingSlash enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// RedirectFixedPath if enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// HandleMethodNotAllowed if enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// ForwardedByClientIP if enabled, client IP will be parsed from the request's headers that
	// match those stored at `(*gin.Engine).RemoteIPHeaders`. If no IP was
	// fetched, it falls back to the IP obtained from
	// `(*gin.Context).Request.RemoteAddr`.
	// ForwardedByClientIP bool

	// AppEngine was deprecated.
	// Deprecated: USE `TrustedPlatform` WITH VALUE `gin.PlatformGoogleAppEngine` INSTEAD
	// #726 #755 If enabled, it will trust some headers starting with
	// 'X-AppEngine...' for better integration with that PaaS.
	// AppEngine bool

	// UseRawPath if enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool

	// UnescapePathValues if true, the path value will be unescaped.
	// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
	// as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool

	// RemoveExtraSlash a parameter can be parsed from the URL even with extra slashes.
	// See the PR #1817 and issue #1644
	RemoveExtraSlash bool

	// RemoteIPHeaders list of headers used to obtain the client IP when
	// `(*gin.Engine).ForwardedByClientIP` is `true` and
	// `(*gin.Context).Request.RemoteAddr` is matched by at least one of the
	// network origins of list defined by `(*gin.Engine).SetTrustedProxies()`.
	// RemoteIPHeaders []string

	// TrustedPlatform if set to a constant of value gin.Platform*, trusts the headers set by
	// that platform, for example to determine the client IP
	// TrustedPlatform string

	// todo: del
	// MaxMultipartMemory value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	// MaxMultipartMemory int64

	// UseH2C enable h2c support.
	UseH2C bool

	// todo: del
	// ContextWithFallback enable fallback Context.Deadline(), Context.Done(), Context.Err() and Context.Value() when Context.Request.Context() is not nil.
	// ContextWithFallback bool

	delims render.Delims
	// secureJSONPrefix string
	// HTMLRender       render.HTMLRender
	// FuncMap        template.FuncMap
	allNoRoute     HandlersChain[T]
	allNoMethod    HandlersChain[T]
	noRoute        HandlersChain[T]
	noMethod       HandlersChain[T]
	pool           sync.Pool
	trees          methodTrees[T]
	maxParams      uint16
	maxSections    uint16
	trustedProxies []string
	// trustedCIDRs     []*net.IPNet
}

var _ IRouter[IContext] = (*Engine[IContext])(nil)

func (engine *Engine[T]) allocateContext(t T, maxParams uint16) T {
	// v := make(Params, 0, maxParams)
	// skippedNodes := make([]skippedNode, 0, engine.maxSections)
	// todo: wait do
	// return &Context{engine: engine, params: &v, skippedNodes: &skippedNodes}

	i := reflect.New(reflect.TypeOf(t).Elem()).Interface().(T)
	// e := *Engine[IContext](engine)
	i.New(engine.maxSections, maxParams)
	return i
}

// func (engine *Engine[T]) allocateContext(t T, maxParams uint16) T {
// 	// v := make(Params, 0, maxParams)
// 	// skippedNodes := make([]skippedNode, 0, engine.maxSections)
// 	// todo: wait do
// 	// return &Context{engine: engine, params: &v, skippedNodes: &skippedNodes}
// 	t.New(engine, maxParams)
// 	return t
// }

// New returns a new blank Engine instance without any middleware attached.
// By default, the configuration is:
// - RedirectTrailingSlash:  true
// - RedirectFixedPath:      false
// - HandleMethodNotAllowed: false
// - ForwardedByClientIP:    true
// - UseRawPath:             false
// - UnescapePathValues:     true
func New[T IContext](t T, opts ...OptionFunc[T]) *Engine[T] {
	debugPrintWARNINGNew()
	engine := &Engine[T]{
		RouterGroup: RouterGroup[T]{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		// FuncMap:                template.FuncMap{},
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: false,
		// ForwardedByClientIP:    true,
		// RemoteIPHeaders:    []string{"X-Forwarded-For", "X-Real-IP"},
		// TrustedPlatform:    defaultPlatform,
		UseRawPath:         false,
		RemoveExtraSlash:   false,
		UnescapePathValues: true,
		// MaxMultipartMemory: defaultMultipartMemory,
		trees:  make(methodTrees[T], 0, 9),
		delims: render.Delims{Left: "{{", Right: "}}"},
		// secureJSONPrefix: "while(1);",
		trustedProxies: []string{"0.0.0.0/0", "::/0"},
		// trustedCIDRs:     defaultTrustedCIDRs,
	}
	engine.RouterGroup.engine = engine
	engine.pool.New = func() any {
		return engine.allocateContext(t, engine.maxParams)
	}
	return engine.With(opts...)
}

// Default returns an Engine instance with the Logger and Recovery middleware already attached.
func Default(opts ...OptionFunc[*Context]) *Engine[*Context] {
	debugPrintWARNINGDefault()
	engine := New(&Context{})
	engine.Use(Logger[*Context](), Recovery[*Context]())
	return engine.With(opts...)
}

func (engine *Engine[T]) Handler() http.Handler {
	if !engine.UseH2C {
		return engine
	}

	h2s := &http2.Server{}
	return h2c.NewHandler(engine, h2s)
}

// Delims sets template left and right delims and returns an Engine instance.
func (engine *Engine[T]) Delims(left, right string) *Engine[T] {
	engine.delims = render.Delims{Left: left, Right: right}
	return engine
}

// // SecureJsonPrefix sets the secureJSONPrefix used in Context.SecureJSON.
// func (engine *Engine[T]) SecureJsonPrefix(prefix string) *Engine[T] {
// 	engine.secureJSONPrefix = prefix
// 	return engine
// }

// note: 不在默认支持模版

// // LoadHTMLGlob loads HTML files identified by glob pattern
// // and associates the result with HTML renderer.
// func (engine *Engine[T]) LoadHTMLGlob(pattern string) {
// 	left := engine.delims.Left
// 	right := engine.delims.Right
// 	templ := template.Must(template.New("").Delims(left, right).Funcs(engine.FuncMap).ParseGlob(pattern))

// 	if IsDebugging() {
// 		debugPrintLoadTemplate(templ)
// 		engine.HTMLRender = render.HTMLDebug{Glob: pattern, FuncMap: engine.FuncMap, Delims: engine.delims}
// 		return
// 	}

// 	engine.SetHTMLTemplate(templ)
// }

// // LoadHTMLFiles loads a slice of HTML files
// // and associates the result with HTML renderer.
// func (engine *Engine[T]) LoadHTMLFiles(files ...string) {
// 	if IsDebugging() {
// 		engine.HTMLRender = render.HTMLDebug{Files: files, FuncMap: engine.FuncMap, Delims: engine.delims}
// 		return
// 	}

// 	templ := template.Must(template.New("").Delims(engine.delims.Left, engine.delims.Right).Funcs(engine.FuncMap).ParseFiles(files...))
// 	engine.SetHTMLTemplate(templ)
// }

// // SetHTMLTemplate associate a template with HTML renderer.
// func (engine *Engine[T]) SetHTMLTemplate(templ *template.Template) {
// 	if len(engine.trees) > 0 {
// 		debugPrintWARNINGSetHTMLTemplate()
// 	}

// 	engine.HTMLRender = render.HTMLProduction{Template: templ.Funcs(engine.FuncMap)}
// }

// SetFuncMap sets the FuncMap used for template.FuncMap.
// func (engine *Engine[T]) SetFuncMap(funcMap template.FuncMap) {
// 	engine.FuncMap = funcMap
// }

// NoRoute adds handlers for NoRoute. It returns a 404 code by default.
func (engine *Engine[T]) NoRoute(handlers ...HandlerFunc[T]) {
	engine.noRoute = handlers
	engine.rebuild404Handlers()
}

// NoMethod sets the handlers called when Engine.HandleMethodNotAllowed = true.
func (engine *Engine[T]) NoMethod(handlers ...HandlerFunc[T]) {
	engine.noMethod = handlers
	engine.rebuild405Handlers()
}

// Use attaches a global middleware to the router. i.e. the middleware attached through Use() will be
// included in the handlers chain for every single request. Even 404, 405, static files...
// For example, this is the right place for a logger or error management middleware.
func (engine *Engine[T]) Use(middleware ...HandlerFunc[T]) IRoutes[T] {
	engine.RouterGroup.Use(middleware...)
	engine.rebuild404Handlers()
	engine.rebuild405Handlers()
	return engine
}

// With returns a Engine with the configuration set in the OptionFunc.
func (engine *Engine[T]) With(opts ...OptionFunc[T]) *Engine[T] {
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

func (engine *Engine[T]) rebuild404Handlers() {
	engine.allNoRoute = engine.combineHandlers(engine.noRoute)
}

func (engine *Engine[T]) rebuild405Handlers() {
	engine.allNoMethod = engine.combineHandlers(engine.noMethod)
}

func (engine *Engine[T]) addRoute(method, path string, handlers HandlersChain[T]) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	debugPrintRoute(method, path, handlers)

	root := engine.trees.get(method)
	if root == nil {
		root = new(node[T])
		root.fullPath = "/"
		engine.trees = append(engine.trees, MethodTree[T]{method: method, root: root})
	}
	root.addRoute(path, handlers)

	if paramsCount := countParams(path); paramsCount > engine.maxParams {
		engine.maxParams = paramsCount
	}

	if sectionsCount := countSections(path); sectionsCount > engine.maxSections {
		engine.maxSections = sectionsCount
	}
}

func (engine *Engine[T]) GetSkippedNodes() *[]SkippedNode[T] {
	skippedNodes := make([]SkippedNode[T], 0, engine.maxSections)

	return &skippedNodes
}

// Routes returns a slice of registered routes, including some useful information, such as:
// the http method, path and the handler name.
func (engine *Engine[T]) Routes() (routes RoutesInfo[T]) {
	for _, tree := range engine.trees {
		routes = iterate("", tree.method, routes, tree.root)
	}
	return routes
}

func iterate[T IContext](path, method string, routes RoutesInfo[T], root *node[T]) RoutesInfo[T] {
	path += root.path
	if len(root.handlers) > 0 {
		handlerFunc := root.handlers.Last()
		routes = append(routes, RouteInfo[T]{
			Method:      method,
			Path:        path,
			Handler:     nameOfFunction(handlerFunc),
			HandlerFunc: handlerFunc,
		})
	}
	for _, child := range root.children {
		routes = iterate(path, method, routes, child)
	}
	return routes
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

// // SetTrustedProxies set a list of network origins (IPv4 addresses,
// // IPv4 CIDRs, IPv6 addresses or IPv6 CIDRs) from which to trust
// // request's headers that contain alternative client IP when
// // `(*gin.Engine).ForwardedByClientIP` is `true`. `TrustedProxies`
// // feature is enabled by default, and it also trusts all proxies
// // by default. If you want to disable this feature, use
// // Engine.SetTrustedProxies(nil), then Context.ClientIP() will
// // return the remote address directly.
// func (engine *Engine[T]) SetTrustedProxies(trustedProxies []string) error {
// 	engine.trustedProxies = trustedProxies
// 	return engine.parseTrustedProxies()
// }

// // isUnsafeTrustedProxies checks if Engine.trustedCIDRs contains all IPs, it's not safe if it has (returns true)
// func (engine *Engine[T]) isUnsafeTrustedProxies() bool {
// 	return engine.isTrustedProxy(net.ParseIP("0.0.0.0")) || engine.isTrustedProxy(net.ParseIP("::"))
// }

// // parseTrustedProxies parse Engine.trustedProxies to Engine.trustedCIDRs
// func (engine *Engine[T]) parseTrustedProxies() error {
// 	trustedCIDRs, err := engine.prepareTrustedCIDRs()
// 	engine.trustedCIDRs = trustedCIDRs
// 	return err
// }

// isTrustedProxy will check whether the IP address is included in the trusted list according to Engine.trustedCIDRs
// func (engine *Engine[T]) isTrustedProxy(ip net.IP) bool {
// 	if engine.trustedCIDRs == nil {
// 		return false
// 	}
// 	for _, cidr := range engine.trustedCIDRs {
// 		if cidr.Contains(ip) {
// 			return true
// 		}
// 	}
// 	return false
// }

// validateHeader will parse X-Forwarded-For header and return the trusted client IP address
// func (engine *Engine[T]) validateHeader(header string) (clientIP string, valid bool) {
// 	if header == "" {
// 		return "", false
// 	}
// 	items := strings.Split(header, ",")
// 	for i := len(items) - 1; i >= 0; i-- {
// 		ipStr := strings.TrimSpace(items[i])
// 		ip := net.ParseIP(ipStr)
// 		if ip == nil {
// 			break
// 		}

// 		// X-Forwarded-For is appended by proxy
// 		// Check IPs in reverse order and stop when find untrusted proxy
// 		if (i == 0) || (!engine.isTrustedProxy(ip)) {
// 			return ipStr, true
// 		}
// 	}
// 	return "", false
// }

// updateRouteTree do update to the route tree recursively
func updateRouteTree[T IContext](n *node[T]) {
	n.path = strings.ReplaceAll(n.path, escapedColon, colon)
	n.fullPath = strings.ReplaceAll(n.fullPath, escapedColon, colon)
	n.indices = strings.ReplaceAll(n.indices, backslash, colon)
	if n.children == nil {
		return
	}
	for _, child := range n.children {
		updateRouteTree(child)
	}
}

// updateRouteTrees do update to the route trees
func (engine *Engine[T]) updateRouteTrees() {
	for _, tree := range engine.trees {
		updateRouteTree(tree.root)
	}
}

// parseIP parse a string representation of an IP and returns a net.IP with the
// minimum byte representation or nil if input is invalid.
func parseIP(ip string) net.IP {
	parsedIP := net.ParseIP(ip)

	if ipv4 := parsedIP.To4(); ipv4 != nil {
		// return ip in a 4-byte representation
		return ipv4
	}

	// return ip in a 16-byte representation or nil
	return parsedIP
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine[T]) Run(addr ...string) (err error) {
	defer func() { debugPrintError(err) }()

	// todo: del
	// if engine.isUnsafeTrustedProxies() {
	// 	debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
	// 		"Please check https://github.com/nbcx/hi/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	// }
	engine.updateRouteTrees()
	address := resolveAddress(addr)
	debugPrint("Listening and serving HTTP on %s\n", address)
	err = http.ListenAndServe(address, engine.Handler())
	return
}

// RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine[T]) RunTLS(addr, certFile, keyFile string) (err error) {
	debugPrint("Listening and serving HTTPS on %s\n", addr)
	defer func() { debugPrintError(err) }()

	// todo: del
	// if engine.isUnsafeTrustedProxies() {
	// 	debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
	// 		"Please check https://github.com/nbcx/hi/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	// }

	err = http.ListenAndServeTLS(addr, certFile, keyFile, engine.Handler())
	return
}

// RunUnix attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified unix socket (i.e. a file).
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine[T]) RunUnix(file string) (err error) {
	debugPrint("Listening and serving HTTP on unix:/%s", file)
	defer func() { debugPrintError(err) }()

	// todo: del
	// if engine.isUnsafeTrustedProxies() {
	// 	debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
	// 		"Please check https://github.com/nbcx/hi/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	// }

	listener, err := net.Listen("unix", file)
	if err != nil {
		return
	}
	defer listener.Close()
	defer os.Remove(file)

	err = http.Serve(listener, engine.Handler())
	return
}

// RunFd attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified file descriptor.
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine[T]) RunFd(fd int) (err error) {
	debugPrint("Listening and serving HTTP on fd@%d", fd)
	defer func() { debugPrintError(err) }()

	// todo: del
	// if engine.isUnsafeTrustedProxies() {
	// 	debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
	// 		"Please check https://github.com/nbcx/hi/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	// }

	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd@%d", fd))
	listener, err := net.FileListener(f)
	if err != nil {
		return
	}
	defer listener.Close()
	err = engine.RunListener(listener)
	return
}

// RunQUIC attaches the router to a http.Server and starts listening and serving QUIC requests.
// It is a shortcut for http3.ListenAndServeQUIC(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine[T]) RunQUIC(addr, certFile, keyFile string) (err error) {
	debugPrint("Listening and serving QUIC on %s\n", addr)
	defer func() { debugPrintError(err) }()

	// todo: del
	// if engine.isUnsafeTrustedProxies() {
	// 	debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
	// 		"Please check https://pkg.go.dev/github.com/nbcx/hi#readme-don-t-trust-all-proxies for details.")
	// }

	err = http3.ListenAndServeQUIC(addr, certFile, keyFile, engine.Handler())
	return
}

// RunListener attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified net.Listener
func (engine *Engine[T]) RunListener(listener net.Listener) (err error) {
	debugPrint("Listening and serving HTTP on listener what's bind with address@%s", listener.Addr())
	defer func() { debugPrintError(err) }()

	// todo: del
	// if engine.isUnsafeTrustedProxies() {
	// 	debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
	// 		"Please check https://github.com/nbcx/hi/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	// }

	err = http.Serve(listener, engine.Handler())
	return
}

// ServeHTTP conforms to the http.Handler interface.
func (engine *Engine[T]) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(T)
	// c.writermem.reset(w)
	// c.Request = req
	c.Init(w, req)

	engine.handleHTTPRequest(c)

	engine.pool.Put(c)
}

// HandleContext re-enters a context that has been rewritten.
// This can be done by setting c.Request.URL.Path to your new target.
// Disclaimer: You can loop yourself to deal with this, use wisely.
func (engine *Engine[T]) HandleContext(c T) {
	oldIndexValue := c.GetExecer().GetIndex()
	c.Reset() // todo: small reset
	engine.handleHTTPRequest(c)

	// c.index = oldIndexValue
	c.GetExecer().SetIndex(oldIndexValue)
}

func (engine *Engine[T]) handleHTTPRequest(c T) {
	req := c.Req()
	httpMethod := req.Method
	rPath := req.URL.Path
	unescape := false
	if engine.UseRawPath && len(req.URL.RawPath) > 0 {
		rPath = req.URL.RawPath
		unescape = engine.UnescapePathValues
	}

	if engine.RemoveExtraSlash {
		rPath = cleanPath(rPath)
	}

	sk := engine.GetSkippedNodes()

	// Find root of the tree for the given HTTP method
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}
		root := t[i].root
		// Find route in tree
		value := root.getValue(rPath, c.GetParams(), sk, unescape)
		if value.params != nil {
			// todo: need check
			// c.Params = *value.params
			c.SetParams(value.params)
		}
		if value.handlers != nil {
			// c.handlers = value.handlers
			// c.fullPath = value.fullPath
			c.SetExecer(NewExecer(c, value.handlers))
			// c.SetHandlers(value.handlers)
			c.SetFullPath(value.fullPath)
			c.Next()
			c.WriterMem().WriteHeaderNow()
			return
		}
		if httpMethod != http.MethodConnect && rPath != "/" {
			if value.tsr && engine.RedirectTrailingSlash {
				redirectTrailingSlash(c)
				return
			}
			if engine.RedirectFixedPath && redirectFixedPath(c, root, engine.RedirectFixedPath) {
				return
			}
		}
		break
	}

	if engine.HandleMethodNotAllowed && len(t) > 0 {
		// According to RFC 7231 section 6.5.5, MUST generate an Allow header field in response
		// containing a list of the target resource's currently supported methods.
		allowed := make([]string, 0, len(t)-1)
		for _, tree := range engine.trees {
			if tree.method == httpMethod {
				continue
			}
			if value := tree.root.getValue(rPath, nil, sk, unescape); value.handlers != nil {
				allowed = append(allowed, tree.method)
			}
		}
		if len(allowed) > 0 {
			// c.handlers = engine.allNoMethod
			c.SetExecer(NewExecer(c, engine.allNoMethod))
			c.WriterMem().Header().Set("Allow", strings.Join(allowed, ", "))
			serveError(c, http.StatusMethodNotAllowed, default405Body)
			return
		}
	}

	// c.handlers = engine.allNoRoute
	c.SetExecer(NewExecer(c, engine.allNoRoute))
	serveError(c, http.StatusNotFound, default404Body)
}

var mimePlain = []string{MIMEPlain}

func serveError[T IContext](c T, code int, defaultMessage []byte) {
	c.WriterMem().status = code
	c.Next()
	if c.WriterMem().Written() {
		return
	}
	if c.WriterMem().Status() == code {
		c.WriterMem().Header()["Content-Type"] = mimePlain
		_, err := c.Rsp().Write(defaultMessage)
		if err != nil {
			debugPrint("cannot write message to writer during serve error: %v", err)
		}
		return
	}
	c.WriterMem().WriteHeaderNow()
}

func redirectTrailingSlash[T IContext](c T) {
	req := c.Req()
	p := req.URL.Path
	if prefix := path.Clean(req.Header.Get("X-Forwarded-Prefix")); prefix != "." {
		prefix = regSafePrefix.ReplaceAllString(prefix, "")
		prefix = regRemoveRepeatedChar.ReplaceAllString(prefix, "/")

		p = prefix + "/" + req.URL.Path
	}
	req.URL.Path = p + "/"
	if length := len(p); length > 1 && p[length-1] == '/' {
		req.URL.Path = p[:length-1]
	}
	redirectRequest(c)
}

func redirectFixedPath[T IContext](c T, root *node[T], trailingSlash bool) bool {
	req := c.Req()
	rPath := req.URL.Path

	if fixedPath, ok := root.findCaseInsensitivePath(cleanPath(rPath), trailingSlash); ok {
		req.URL.Path = bytesconv.BytesToString(fixedPath)
		redirectRequest(c)
		return true
	}
	return false
}

func redirectRequest[T IContext](c T) {
	req := c.Req()
	rPath := req.URL.Path
	rURL := req.URL.String()

	code := http.StatusMovedPermanently // Permanent redirect, request with GET method
	if req.Method != http.MethodGet {
		code = http.StatusTemporaryRedirect
	}
	debugPrint("redirecting request %d: %s --> %s", code, rPath, rURL)
	http.Redirect(c.Rsp(), req, rURL, code)
	c.WriterMem().WriteHeaderNow()
}
