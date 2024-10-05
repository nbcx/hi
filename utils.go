// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package hi

import (
	"encoding/xml"
	"errors"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"unicode"
)

// BindKey indicates a default bind key.
const BindKey = "_gin-gonic/gin/bindkey"

// Bind is a helper function for given interface object and returns a Gin middleware.
// func Bind[T IContext](val any) HandlerFunc[T] {
// 	value := reflect.ValueOf(val)
// 	if value.Kind() == reflect.Ptr {
// 		panic(`Bind struct can not be a pointer. Example:
// 	Use: gin.Bind(Struct{}) instead of gin.Bind(&Struct{})
// `)
// 	}
// 	typ := value.Type()

// 	return func(c T) {
// 		obj := reflect.New(typ).Interface()
// 		if c.Bind(obj) == nil {
// 			c.Set(BindKey, obj)
// 		}
// 	}
// }

// WrapF is a helper function for wrapping http.HandlerFunc and returns a Gin middleware.
func WrapF[T IContext](f http.HandlerFunc) HandlerFunc[T] {
	return func(c T) {
		f(c.Rsp(), c.Req())
	}
}

// WrapH is a helper function for wrapping http.Handler and returns a Gin middleware.
func WrapH[T IContext](h http.Handler) HandlerFunc[T] {
	return func(c T) {
		h.ServeHTTP(c.Rsp(), c.Req())
	}
}

// H is a shortcut for map[string]any
type H map[string]any

// MarshalXML allows type H to be used with xml.Marshal.
func (h H) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

func chooseData(custom, wildcard any) any {
	if custom != nil {
		return custom
	}
	if wildcard != nil {
		return wildcard
	}
	panic("negotiation config is invalid")
}

func parseAccept(acceptHeader string) []string {
	parts := strings.Split(acceptHeader, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if i := strings.IndexByte(part, ';'); i > 0 {
			part = part[:i]
		}
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func nameOfFunction(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}

// https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// todo: ClientIP() string // todo: 待移除
// GetIP returns request real ip.

func ClientIP(r *http.Request) (ip string) {
	ip, _ = ClientIPE(r)
	return
}

func ClientIPE(r *http.Request) (string, error) {
	// X-Real-IP：只包含客户端机器的一个IP，如果为空，某些代理服务器（如Nginx）会填充此header。
	// X-Forwarded-For：一系列的IP地址列表，以,分隔，每个经过的代理服务器都会添加一个IP。
	// RemoteAddr：包含客户端的真实IP地址。 这是Web服务器从其接收连接并将响应发送到的实际物理IP地址。 但是，如果客户端通过代理连接，它将提供代理的IP地址。
	// RemoteAddr是最可靠的，但是如果客户端位于代理之后或使用负载平衡器或反向代理服务器时，它将永远不会提供正确的IP地址，因此顺序是先是X-REAL-IP，然后是X-FORWARDED-FOR，然后是 RemoteAddr。 请注意，恶意用户可以创建伪造的X-REAL-IP和X-FORWARDED-FOR标头。
	ip := r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	ip = r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	return "", errors.New("no valid ip found")
}
