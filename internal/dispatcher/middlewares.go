package dispatcher

import (
	"bytes"
	"fmt"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"io/ioutil"
)

// RecoverMiddleware
func (d *Dispatcher) RecoverMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					d.L().Critical("[PANIC RECOVER] %s", logger.Args(err.Error()), logger.Stack("stacktrace"))
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}

// RawBodyPreMiddleware
func (d *Dispatcher) RawBodyPreMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		buf, _ := ioutil.ReadAll(c.Request().Body)
		rdr := ioutil.NopCloser(bytes.NewBuffer(buf))
		c.Request().Body = rdr
		common.SetRawBodyContext(c, buf)
		return next(c)
	}
}

// BodyDumpMiddleware
func (d *Dispatcher) BodyDumpMiddleware() echo.MiddlewareFunc {
	return middleware.BodyDump(func(ctx echo.Context, reqBody, resBody []byte) {
		data := map[string]interface{}{
			"request_headers":  common.RequestResponseHeadersToString(ctx.Request().Header),
			"request_body":     string(reqBody),
			"response_headers": common.RequestResponseHeadersToString(ctx.Response().Header()),
			"response_body":    string(resBody),
		}
		d.L().Info(ctx.Path(), logger.WithFields(data))
	})
}
