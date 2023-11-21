package context

import (
	"errors"

	"github.com/labstack/echo/v4"
)

type Context struct {
	echo.Context
	*Components
}

var (
	ErrorCastFail = errors.New("failed to cast the context")
)

func NewContext(e echo.Context, c *Components) *Context {
	return &Context{
		Context:    e,
		Components: c,
	}
}

func FromEchoContext(e echo.Context) (*Context, error) {
	ctx, ok := e.(*Context)
	if !ok {
		return nil, ErrorCastFail
	}
	return ctx, nil
}
