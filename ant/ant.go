package ant

import (
	"log"
	"net/http"
	"strings"
)

type Engine struct {
	RouterGroup
	router    *Router
	groupList []*RouterGroup
}

type RouterGroup struct {
	prefix         string
	engine         *Engine
	middlewareList []HandlerFunc
}

func (g *RouterGroup) Get(path string, f HandlerFunc) {
	g.addRoute("GET", path, f)
}

func (g *RouterGroup) Post(path string, f HandlerFunc) {
	g.addRoute("POST", path, f)
}

// Add middleware to the group
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewareList = append(group.middlewareList, middlewares...)
}

func (g *RouterGroup) addRoute(method string, path string, f HandlerFunc) {
	pattern := NormalizePath(g.prefix + "/" + path)
	log.Printf("route registered: %4s /%s", method, pattern)
	g.engine.router.AddRoute(method, pattern, f)
}

func (g *RouterGroup) AddGroup(groupName string) *RouterGroup {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/"
	if strings.IndexFunc(groupName, func(r rune) bool {
		if strings.IndexRune(chars, r) < 0 {
			return true
		}
		return false
	}) >= 0 {
		panic("invalid groupName: " + groupName)
	}

	engine := g.engine
	group := RouterGroup{
		prefix: NormalizePath(g.prefix + "/" + groupName),
		engine: engine,
	}
	engine.groupList = append(engine.groupList, &group)
	return &group
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	context := newContext(writer, request)

	// find middlewares of groups
	path := NormalizePath(request.URL.Path)
	handlerList := make([]HandlerFunc, 0)
	for _, group := range e.groupList {
		if strings.HasPrefix(path, group.prefix) {
			handlerList = append(handlerList, group.middlewareList...)
		}
	}
	context.handlerList = handlerList

	context.engine = e
	e.router.handle(context)
}

func newContext(writer http.ResponseWriter, request *http.Request) *Context {
	return &Context{
		Writer:  writer,
		Request: request,
		Method:  request.Method,
		Path:    request.URL.Path,
		index:   -1,
	}
}

func (e *Engine) Run(addr string) {
	log.Printf("ant serve at %s\n", addr)
	http.ListenAndServe(addr, e)
}

func (e *Engine) RunAsync(addr string, isStarted chan bool) {
	go http.ListenAndServe(addr, e)
	isStarted <- true
}

func NewEngine() *Engine {
	e := Engine{}
	e.engine = &e
	e.router = NewRouter()
	e.groupList = append(
		e.groupList,
		&e.RouterGroup,
	)
	return &e
}

func Default() *Engine {
	e := NewEngine()
	e.Use(Recovery())
	return e
}
