package ant

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HandlerFunc defines the request handler used by ant engine
type HandlerFunc func(*Context)

type M map[string]any

type Context struct {
	Writer      http.ResponseWriter
	Request     *http.Request
	Method      string
	Path        string
	Param       map[string]string
	StatusCode  int
	handlerList []HandlerFunc
	index       int
	engine      *Engine
}

// execute handlers from first to last
func (c *Context) Next() {
	c.index++
	len := len(c.handlerList)
	for c.index < len {
		c.handlerList[c.index](c)
		c.index++
	}
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) String(code int, format string, values ...any) {
	c.SetHeader("content-type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) HTML(code int, htmlStr string) {
	c.SetHeader("content-type", "text/html")
	c.Status(code)
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(htmlStr))
}

func (c *Context) JSON(code int, jsonObj any) {
	c.SetHeader("content-type", "text/plain")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(jsonObj); err != nil {
		panic(err)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) QueryValue(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) ParseJSON(target any) {
	//all, _ := ioutil.ReadAll(c.Request.Body)
	//log.Println(string(all))
	encoder := json.NewDecoder(c.Request.Body)
	if err := encoder.Decode(target); err != nil {
		panic(err)
	}
}

// skip remaining handlers
func (c *Context) Fail(statusCode int, message string) {
	c.index = len(c.handlerList)
	c.String(statusCode, message)
}

func (c *Context) InternalError() {
	c.String(http.StatusInternalServerError, "system error")
}

func (c *Context) Use(middlewares ...HandlerFunc) {
	c.handlerList = append(c.handlerList, middlewares...)
}
