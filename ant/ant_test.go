package ant

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

var addr = "localhost:8881"
var hello = "hello world"

type RegisterReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

func TestMain(m *testing.M) {
	isStarted := make(chan bool, 1)
	startServer(addr, isStarted)
	<-isStarted

	os.Exit(m.Run())
}

func startServer(addr string, isStarted chan bool) {
	hello = strings.ToLower(hello)
	engine := Default()
	// config
	engine.Get("/", func(c *Context) {
		c.String(http.StatusOK, "hello world")
	})
	engine.Get("/html", func(c *Context) {
		c.HTML(http.StatusOK, "<h1>Hello World</h1>")
	})
	engine.Get("/json", func(c *Context) {
		c.JSON(http.StatusOK, M{
			"path":   c.Path,
			"method": c.Method,
		})
	})
	engine.Post("/login", func(c *Context) {
		username := c.FormValue("username")
		password := c.FormValue("password")
		c.JSON(http.StatusOK, M{
			"username":   username,
			"password":   password,
			"login_time": time.Now(),
		})
	})
	engine.Get("/error", func(c *Context) {
		panic("test error")
	})

	group := engine.AddGroup("/g1")
	group.Post("/register", func(c *Context) {
		req := RegisterReq{}
		c.ParseJSON(&req)
		c.JSON(http.StatusOK, req)
	})

	group = engine.AddGroup("/g2")
	group.Use(func(context *Context) {
		hello = strings.ToUpper(hello)
	})
	group.Get("/hello", func(c *Context) {
		c.String(http.StatusOK, "hello group")
	})

	engine.RunAsync(addr, isStarted)
}

func getURL(addr string, path string) string {
	path = NormalizePath(path)
	return fmt.Sprintf("http://%s/%s", addr, path)
}

func parseJSON(resp *http.Response) (m map[string]string) {
	m = make(map[string]string)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&m)
	return
}

func TestEngine(t *testing.T) {
	resp, err := http.Get(getURL(addr, "/"))
	assert.Nil(t, err, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "hello world", string(body))

	resp, err = http.Get(getURL(addr, "/html"))
	assert.Nil(t, err, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "<h1>Hello World</h1>", string(body))
	assert.Equal(t, "text/html", resp.Header.Get("content-type"))

	resp, err = http.Get(getURL(addr, "/json"))
	assert.Nil(t, err, err)
	m := parseJSON(resp)
	assert.Equal(t, "/json", m["path"])

	resp, err = http.PostForm(getURL(addr, "/login"), url.Values{"username": {"zhangsan"}, "password": {"pwd"}})
	assert.Nil(t, err, err)
	m = parseJSON(resp)
	assert.Equal(t, "zhangsan", m["username"])
	assert.Equal(t, "pwd", m["password"])

	resp, err = http.Get(getURL(addr, "/error"))
	assert.Nil(t, err, err)
	m = parseJSON(resp)
	assert.Equal(t, "test error", m["error"])

	resp, err = http.Get(getURL(addr, "/not-exists"))
	assert.Nil(t, err, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	assert.Equal(t, "hello world", hello)
	resp, err = http.Get(getURL(addr, "/g2/hello"))
	assert.Nil(t, err, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "hello group", string(body))
	assert.Equal(t, "HELLO WORLD", hello)
}
