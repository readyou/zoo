package ant

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestContext_Next(t *testing.T) {
	c := Context{
		index: -1,
	}
	b := strings.Builder{}
	a := strings.Builder{}
	for i := 1; i < 4; i++ {
		j := i
		c.Use(func(context *Context) {
			b.WriteString(fmt.Sprintf("B%d", j))
			context.Next()
			a.WriteString(fmt.Sprintf("A%d", j))
		})
	}
	c.Next()
	assert.Equal(t, b.String(), "B1B2B3")
	assert.Equal(t, a.String(), "A3A2A1")
}

func TestContext_Next_justHandleBefore(t *testing.T) {
	c := Context{
		index: -1,
	}
	b := strings.Builder{}
	for i := 1; i < 4; i++ {
		j := i
		c.Use(func(context *Context) {
			b.WriteString(fmt.Sprintf("B%d", j))
		})
	}
	c.Next()
	assert.Equal(t, b.String(), "B1B2B3")
}

func TestContext_Next_justHandleAfter(t *testing.T) {
	c := Context{
		index: -1,
	}
	a := strings.Builder{}
	for i := 1; i < 4; i++ {
		j := i
		c.Use(func(context *Context) {
			context.Next()
			a.WriteString(fmt.Sprintf("A%d", j))
		})
	}
	c.Next()
	assert.Equal(t, a.String(), "A3A2A1")
}
