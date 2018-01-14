package modules

import (
	"fmt"
	"strings"

	"github.com/evilsocket/bettercap-ng/log"

	"github.com/gin-gonic/gin"
	"gopkg.in/unrolled/secure.v1"
)

func SecurityMiddleware() gin.HandlerFunc {
	rules := secure.New(secure.Options{
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		ReferrerPolicy:     "same-origin",
	})

	return func(c *gin.Context) {
		err := rules.Process(c.Writer, c.Request)
		if err != nil {
			who := strings.Split(c.Request.RemoteAddr, ":")[0]
			req := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
			log.Warning("%s > %s | Security exception: %s", who, req, err)
			c.Abort()
			return
		}

		if status := c.Writer.Status(); status > 300 && status < 399 {
			c.Abort()
		}
	}
}
