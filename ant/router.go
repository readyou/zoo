package ant

import (
	"net/http"
	"strings"
)

// use trie to match request's route
type Router struct {
	rootMap    map[string]*node       // map of <method, root>
	handlerMap map[string]HandlerFunc // map of <path, handler>
}

type node struct {
	pattern   string
	part      string
	childList []*node
	isWild    bool
	isRoute   bool
}

func NewRouter() *Router {
	return &Router{
		rootMap:    make(map[string]*node),
		handlerMap: make(map[string]HandlerFunc),
	}
}

func (r *Router) AddRoute(method string, pattern string, f HandlerFunc) {
	pattern = NormalizePath(pattern)

	n, ok := r.rootMap[method]
	if !ok {
		n = &node{}
		r.rootMap[method] = n
	}
	parts := parseParts(pattern)
	n.insert(pattern, parts, 0)

	key := routeKey(method, pattern)
	r.handlerMap[key] = f
}

func routeKey(method, pattern string) string {
	return method + "_" + pattern
}

func parseParts(pattern string) (partList []string) {
	strs := strings.Split(pattern, "/")
	for _, v := range strs {
		partList = append(partList, v)
		// only reserve one '*', discard following parts
		if len(v) > 0 && v[0] == '*' {
			break
		}
	}
	return partList
}

func NormalizePath(path string) string {
	path = strings.ReplaceAll(path, "//", "/")
	if strings.Index(path, ":/") >= 0 {
		panic("invalid route pattern: " + path)
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}

func (n *node) insert(pattern string, partList []string, depth int) {
	if depth == len(partList) {
		n.pattern = pattern
		n.isRoute = true
		return
	}
	part := partList[depth]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part: part,
		}
		if len(part) > 0 && (part[0] == '*' || part[0] == ':') {
			child.isWild = true
		}
		n.childList = append(n.childList, child)
	}
	child.insert(pattern, partList, depth+1)
}

func (n *node) matchChild(part string) *node {
	for _, child := range n.childList {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

func (n *node) matchChildren(part string) (nodeList []*node) {
	for _, child := range n.childList {
		if child.part == part || child.isWild {
			nodeList = append(nodeList, child)
		}
	}
	return nodeList
}

func (r *Router) getRoute(method, requestPath string) (f HandlerFunc, params map[string]string) {
	root := r.rootMap[method]
	if root == nil {
		return
	}

	path := NormalizePath(requestPath)
	parts := parseParts(path)
	if n := root.search(parts, 0); n != nil {
		params = make(map[string]string)
		patternParts := parseParts(n.pattern)
		for i, part := range patternParts {
			if part == "" {
				continue
			}
			switch part[0] {
			case ':':
				params[part[1:]] = parts[i]
			case '*':
				if len(part) > 1 {
					params[part[1:]] = strings.Join(parts[i:], "/")
				}
			}
		}

		key := routeKey(method, n.pattern)
		f = r.handlerMap[key]
		return
	}

	return
}

func (r *Router) handle(context *Context) {
	handler, params := r.getRoute(context.Method, context.Path)
	if handler == nil {
		context.String(http.StatusNotFound, "404 NOT FOUND: %s %s\n", context.Method, context.Path)
		return
	}
	context.handlerList = append(context.handlerList, handler)
	context.Param = params
	context.Next()
}

func (n *node) search(partList []string, depth int) *node {
	if depth == len(partList) || strings.HasPrefix(n.part, "*") {
		if n.isRoute {
			return n
		}
		return nil
	}
	part := partList[depth]
	for _, child := range n.matchChildren(part) {
		if ret := child.search(partList, depth+1); ret != nil {
			return ret
		}
	}
	return nil
}
