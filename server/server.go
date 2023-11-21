package server

import (
	"lift/server/context"
	"lift/server/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Server struct {
	echo       *echo.Echo
	components *context.Components
	listenAt   string
}

func NewServer(
	e *echo.Echo,
	c *context.Components,
	listenAt string,
) *Server {
	e.Validator = validator.NewValidator()
	return &Server{
		echo:       e,
		components: c,
		listenAt:   listenAt,
	}
}

func (s *Server) ConvertContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.NewContext(c, s.components)
		return next(ctx)
	}
}

func (s *Server) Run() {
	s.echo.Use(s.ConvertContext)
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.Logger())

	s.echo.Logger.SetLevel(log.INFO)
	s.echo.Logger.Fatal(s.echo.Start(s.listenAt))
}
