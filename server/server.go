package server

import (
	"lift/server/context"
	"lift/server/handlers"
	"lift/server/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type ServerParams struct {
	listenAt string
	logLevel log.Lvl
}

func NewServerParams(listenAt string, logLevel log.Lvl) *ServerParams {
	return &ServerParams{
		listenAt: listenAt,
		logLevel: logLevel,
	}
}

type Server struct {
	echo       *echo.Echo
	components *context.Components
	params     *ServerParams
	errCh      chan error
}

func NewServer(
	e *echo.Echo,
	c *context.Components,
	p *ServerParams,
) *Server {
	e.Validator = validator.NewValidator()
	return &Server{
		echo:       e,
		components: c,
		params:     p,
		errCh:      make(chan error),
	}
}

func (s *Server) ConvertContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.NewContext(c, s.components)
		return next(ctx)
	}
}

func (s *Server) Run() <-chan error {
	s.echo.Use(s.ConvertContext)
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.Logger())

	s.echo.GET("/", handlers.Root)
	s.echo.GET("/connect/:id", handlers.Connect)

	s.echo.Logger.SetLevel(s.params.logLevel)
	go s.start()
	return s.errCh
}

func (s *Server) start() {
	err := s.echo.Start(s.params.listenAt)
	s.errCh <- err
}
