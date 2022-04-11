package ant

import (
	"fmt"
	"log"
	"net/http"
	"zoo/util"
)

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", util.Trace(message))
				c.Fail(http.StatusInternalServerError, message)
			}
		}()

		c.Next()
	}
}
