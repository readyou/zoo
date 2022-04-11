package ant

import (
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"testing"
	"time"
)

func fun(c *Context) {
	c.JSON(http.StatusOK, M{
		"path":   c.Path,
		"method": c.Method,
	})
}

type Param struct {
	method, pattern, path string
	params                map[string]any
}

func TestRouter_NormalizePath(t *testing.T) {
	paramList := [][]string{
		{"/", ""},
		{"//abc", "abc"},
		{"//abc//def", "abc/def"},
		{"/abc", "abc"},
		{"abc/", "abc"},
		{"/abc/", "abc"},
	}
	for _, param := range paramList {
		path := NormalizePath(param[0])
		assert.Equal(t, param[1], path)
	}
}

func TestRouter_AddAndSearch(t *testing.T) {
	router := NewRouter()

	paramList := []Param{
		{"GET", "/", "/", M{}},
		{"POST", "/login", "/login/", M{}},
		{"GET", "/history/:username", "/history/spsp", M{"username": "spsp"}},
		{"GET", "/history/:username/:historyId", "/history/spsp/1", M{"username": "spsp", "historyId": "1"}},
		{"GET", "/dashboard/*board", "/dashboard/board1/sub", M{"board": "board1/sub"}},
	}

	for _, param := range paramList {
		router.AddRoute(param.method, param.pattern, fun)
	}

	n := 100000
	start := time.Now()
	for i := 0; i < n; i++ {
		for _, param := range paramList {
			handlerFunc, p := router.getRoute(param.method, param.path)
			assert.NotNil(t, handlerFunc)
			if len(param.params) > 0 {
				equal := equalMap(param.params, p)
				assert.True(t, equal)
			}
		}
	}
	end := time.Now()
	seconds := end.Sub(start).Seconds()
	log.Printf("Qps: %f\n", float64(n*len(paramList))/seconds)
}

func equalMap(a map[string]any, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v.(string) {
			return false
		}
	}
	return true
}
