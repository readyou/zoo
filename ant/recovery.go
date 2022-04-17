package ant

import (
	"fmt"
	"git.garena.com/xinlong.wu/zoo/util"
	"log"
	"net/http"
)

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", util.Err.Trace(message))
				c.Fail(http.StatusInternalServerError, message)
			}
		}()

		c.Next()
	}
}
