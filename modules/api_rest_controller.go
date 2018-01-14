package modules

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/evilsocket/bettercap-ng/session"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type CommandRequest struct {
	Command string `json:"cmd"`
}

type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
}

func SafeBind(c *gin.Context, obj interface{}) error {
	decoder := json.NewDecoder(io.LimitReader(c.Request.Body, 100*1024))
	if binding.EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if err := decoder.Decode(obj); err != nil {
		return err
	}

	if binding.Validator == nil {
		return nil
	}
	return binding.Validator.ValidateStruct(obj)
}

func BadRequest(c *gin.Context, opt_msg ...string) {
	msg := "Bad Request"
	if len(opt_msg) > 0 {
		msg = opt_msg[0]
	}
	c.JSON(400, ApiResponse{
		Success: false,
		Message: msg,
	})
	c.Abort()
}

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
		c.JSON(200, ApiResponse{Success: true})
	}
}

func ShowRestEvents(c *gin.Context) {
	var err error

	events := session.I.Events.Events()
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
