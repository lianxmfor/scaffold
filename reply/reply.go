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
        if data == nil {
                data = make(map[string]interface{})
        }
        return func(context *gin.Context) {
                if _, exist := data["code"]; !exist {
                        data["code"] = code
                }
                context.JSON(http.StatusOK, data)
        }
}

func ErrorWithMessage(err error, msg string) gin.HandlerFunc {
        return Err(errors.WithMessage(err, msg))
}

func Err(err error) gin.HandlerFunc {
        return func(c *gin.Context) {
                fmt.Printf("%+v\n", err)
                c.JSON(http.StatusBadRequest, Response{
                        Code: http.StatusBadRequest,
                        Msg:  err.Error(),
                })
        }
}
