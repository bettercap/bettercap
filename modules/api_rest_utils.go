package modules

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type CommandRequest struct {
	Command string `json:"cmd"`
}

type APIResponse struct {
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

func BadRequest(c *gin.Context, optMsg ...string) {
	msg := "Bad Request"
	if len(optMsg) > 0 {
		msg = optMsg[0]
	}
	c.JSON(400, APIResponse{
		Success: false,
		Message: msg,
	})
	c.Abort()
}
