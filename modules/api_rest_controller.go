package modules

import (
	"strconv"

	"github.com/evilsocket/bettercap-ng/session"

	"github.com/gin-gonic/gin"
)

func ShowRestSession(c *gin.Context) {
	c.JSON(200, session.I)
}

func RunRestCommand(c *gin.Context) {
	var err error
	var cmd CommandRequest

	if err = SafeBind(c, &cmd); err != nil {
		BadRequest(c)
	}

	err = session.I.Run(cmd.Command)
	if err != nil {
		BadRequest(c, err.Error())
	} else {
		c.JSON(200, APIResponse{Success: true})
	}
}

func ShowRestEvents(c *gin.Context) {
	var err error

	events := session.I.Events.Sorted()
	nmax := len(events)
	n := nmax

	q := c.Request.URL.Query()
	vals := q["n"]
	if len(vals) > 0 {
		n, err = strconv.Atoi(q["n"][0])
		if err == nil {
			if n > nmax {
				n = nmax
			}
		} else {
			n = nmax
		}
	}

	c.JSON(200, events[0:n])
}

func ClearRestEvents(c *gin.Context) {
	session.I.Events.Clear()
	session.I.Events.Add("sys.log.cleared", nil)
	c.JSON(200, gin.H{"success": true})
}
