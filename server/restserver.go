/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package server

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"github.com/azzill/goze/common"
	"github.com/azzill/goze/log"
	"github.com/azzill/goze/midware"
	"github.com/azzill/goze/sql"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type RequestHandler func(ctx *common.RequestCtx) interface{}

type HttpConfig struct {
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}
type RestServer struct {
	controller *RestController
	config     *HttpConfig
	address    string
}

func NewRestServer(address string, config *HttpConfig) *RestServer {
	return &RestServer{controller: &RestController{responseWrapper: list.New()}, config: config, address: address}
}

// customize response by handle it manually return true if handled
type ResponseWrapper interface {
	Wrap(v interface{}, wr http.ResponseWriter) bool
}

type RestController struct {
	mapping            map[RequestMethod]*prefixNode
	requestInterceptor midware.InterceptorChain
	responseWrapper    *list.List
	sql                *sql.SQL
}

type RequestMethod string

var logger = log.NewLogger("RestServer")

const (
	Get     RequestMethod = "GET"
	Post    RequestMethod = "POST"
	Put     RequestMethod = "PUT"
	Delete  RequestMethod = "DELETE"
	Head    RequestMethod = "HEAD"
	Options RequestMethod = "OPTIONS"
)

type RequestMapping struct {
	method  RequestMethod
	depth   int
	pattern string
	handler RequestHandler
}

type prefixNode struct {
	mapped      bool
	matchAll    bool
	prefix      string
	placeholder string
	handler     RequestHandler
	parent      *prefixNode
	children    map[string]*prefixNode
}

var urlFormatRegexp []*regexp.Regexp

func init() {
	r1, _ := regexp.Compile("\\.+/")
	r2, _ := regexp.Compile("/{2,}")
	urlFormatRegexp = []*regexp.Regexp{r1, r2}
}

func (s *RestServer) GET(pattern string, handler RequestHandler) *RestServer {
	return s.Mapping(Get, pattern, handler)
}
func (s *RestServer) POST(pattern string, handler RequestHandler) *RestServer {
	return s.Mapping(Post, pattern, handler)
}
func (s *RestServer) DELETE(pattern string, handler RequestHandler) *RestServer {
	return s.Mapping(Delete, pattern, handler)
}
func (s *RestServer) PUT(pattern string, handler RequestHandler) *RestServer {
	return s.Mapping(Put, pattern, handler)
}

func (s *RestServer) Mapping(method RequestMethod, pattern string, handler RequestHandler) *RestServer {

	if s.controller.mapping == nil {
		s.controller.mapping = map[RequestMethod]*prefixNode{
			Get:     {prefix: "", children: map[string]*prefixNode{}},
			Put:     {prefix: "", children: map[string]*prefixNode{}},
			Post:    {prefix: "", children: map[string]*prefixNode{}},
			Delete:  {prefix: "", children: map[string]*prefixNode{}},
			Head:    {prefix: "", children: map[string]*prefixNode{}},
			Options: {prefix: "", children: map[string]*prefixNode{}}}

	}
	// format pattern
	pattern = strings.TrimSpace(pattern)
	if len(pattern) == 0 {
		panic("empty pattern")
	}
	if pattern[0] == '/' {
		if len(pattern) == 1 {
			pattern = ""
		} else {
			pattern = pattern[1:]
		}
	}
	prefixes := strings.Split(pattern, "/")
	currentNode := s.controller.mapping[method]
	for i := 0; pattern != "" && i < len(prefixes); i++ {
		prefix := prefixes[i]
		placeholder := ""

		//node already match all patterns
		if currentNode.matchAll {
			panic("ambiguous mapping")
		}

		if len(prefix) > 0 && prefix[0] == ':' {
			placeholder = prefix[1:]
			prefix = "*"
		} else if len(prefix) > 0 && prefix[0] == '*' {
			placeholder = prefix[1:]
			prefix = "*"
			if i < len(prefix)-1 {
				logger.Error(pattern, "full pattern placeholder must be the end")
				return s
			}

			if currentNode.children[prefix] != nil {
				logger.Error(pattern, "ambiguous mapping")
				return s
			}

			nextNode := &prefixNode{prefix: prefix, placeholder: placeholder,
				mapped: false, children: nil, matchAll: true}
			currentNode.children[prefix] = nextNode
			currentNode = nextNode
			break

		}

		nextNode := currentNode.children[prefix]

		//allocate children node
		if nextNode == nil {
			nextNode = &prefixNode{prefix: prefix, placeholder: placeholder,
				mapped: false, children: map[string]*prefixNode{}, matchAll: false}
			currentNode.children[prefix] = nextNode
		}
		currentNode = nextNode
	}

	if currentNode.mapped {
		logger.Error(pattern, "ambiguous mapping")
	}

	currentNode.mapped = true
	currentNode.handler = handler
	logger.Info("URL Mapped", method, "/"+pattern)
	return s
}

func (c *RestServer) WithSQL(sql *sql.SQL) {
	c.controller.sql = sql
}

func (c *RestController) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	var obj interface{} = nil
	pv := make(map[string]string)
	url := r.URL.Path

	for _, reg := range urlFormatRegexp {
		url = reg.ReplaceAllString(url, "/")
	}

	println("Formatted: ", url)

	split := strings.Split(url[1:], "/")
	currentNode := c.mapping[RequestMethod(r.Method)]

	for i := 0; url[1:] != "" && i < len(split); i++ {
		s := split[i]
		node := currentNode.children[s]

		if node == nil {
			node = currentNode.children["*"]
			if node == nil {
				currentNode = nil
				break
			}
			if node.matchAll {
				pv[node.placeholder] = strings.Join(split[i:], "/")
				currentNode = node
				break
			} else {
				pv[node.placeholder] = s
			}
		}
		currentNode = node
	}

	//recover from any exception
	defer func() {
		if err := recover(); err != nil {
			HttpError(wr, http.StatusInternalServerError, fmt.Sprint(err), true)
		}
	}()

	//mapped
	if currentNode != nil && currentNode.mapped {

		//Begin sql transaction
		ctx := common.NewRequestCtx(r.URL.Query(), pv, r, r.MultipartForm, wr, c.sql)

		var intercepted bool
		// firstly, handle with interceptor
		intercepted, obj = c.requestInterceptor.CallInterceptors(ctx)

		if !intercepted {
			obj = currentNode.handler(ctx)
		}

		//returned value is not an error commit sql transaction
		if _, ok := obj.(error); !ok {
			if e := ctx.Tx.Commit(); e != nil {
				obj = e
			}
		} else {
			if e := ctx.Tx.Rollback(); e != nil {
				obj = e
			}
		}

		//find a proper response wrapper
		for e := c.responseWrapper.Front(); e != nil; e = e.Next() {
			if e.Value.(ResponseWrapper).Wrap(obj, wr) {
				return
			}
		}

		//default wrapper
		c.defaultResponseWrapper(obj, wr)
	} else { //unmapped
		http.NotFound(wr, r)
		return
	}

}

// append to the top of response
func (s *RestServer) AddResponseWrapper(wrapper ResponseWrapper) {
	s.controller.responseWrapper.PushFront(wrapper)
}

//default response wrapper
func (c *RestController) defaultResponseWrapper(v interface{}, wr http.ResponseWriter) bool {
	var err error

	//Nil
	if v == nil {
		return true
	}

	//Unhandled error
	if e, ok := v.(error); ok {
		HttpError(wr, http.StatusInternalServerError, e.Error(), false)
		return true
	}

	switch (v).(type) {
	case string:
		wr.Header().Set("Content-Type", "text/plain")
		_, err = wr.Write([]byte((v).(string))) //return origin value if string
	default:
		wr.Header().Set("Content-Type", "application/json")
		if bytes, err := json.Marshal(v); err != nil {
			http.Error(wr, err.Error(), 500)
		} else {
			_, err = wr.Write(bytes)
		}
	}

	if err != nil {
		logger.Error("IO Error occurs at response -", err.Error())
	}
	return true
}

func (s *RestServer) StartServer() {
	s.startWith(true)
}

func (s *RestServer) StartServerAsync() *http.Server {
	return s.startWith(false)
}

func (s *RestServer) AddInterceptor(interceptor midware.Interceptor) {
	s.controller.requestInterceptor.AddInterceptor(interceptor)
}
func (s *RestServer) startWith(block bool) *http.Server {

	server := &http.Server{Addr: s.address, Handler: s.controller}

	go func() {
		logger.Info("Goze Server started at", s.address)
		err := server.ListenAndServe()
		logger.Warn(err)

	}()

	if !block {
		return server
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	<-sig

	c, _ := context.WithCancel(context.Background())
	_ = server.Shutdown(c)
	return nil
}

const errorPage = "<h1>%v %v</h1><h2>%v</h2><p>%v</p>"

func HttpError(wr http.ResponseWriter, status int, info string, showtrace bool) {
	wr.Header().Set("Content-Type", "text/html; charset=utf-8")
	wr.Header().Set("X-Content-Type-Options", "nosniff")
	wr.WriteHeader(status)

	callers := ""
	if showtrace {
		for i := 3; true; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			callers = callers + fmt.Sprintf("<div>%v:%v</div>\n", file, line)
		}
	}
	html := fmt.Sprintf(errorPage, status, http.StatusText(status), template.HTMLEscapeString(info), callers)
	_, _ = fmt.Fprintln(wr, html)
}
