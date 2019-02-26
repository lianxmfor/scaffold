package reply

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func Wrap(f func(c *gin.Context) gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := f(c)
		response(c)
	}
}

func Success(code int, data map[string]interface{}) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.JSON(http.StatusBadRequest, Response{
			Code: code,
			Data: data,
		})
	}
}

func ErrorWithMessage(err error, msg string) gin.HandlerFunc {
	return Err(msg, errors.WithMessage(err, msg))
}

func Err(msg string, err error) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("%+v", err)
		c.JSON(http.StatusBadRequest, Response{
			Code: http.StatusBadRequest,
			Msg:  msg,
		})
	}
}
