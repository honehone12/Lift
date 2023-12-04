package errres

import (
	"lift/logger"
	"net/http"

	"github.com/labstack/echo/v4"
)

func BadRequest(err error, l logger.Logger) error {
	l.Warn(err)
	return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
}

func ServerError(err error, l logger.Logger) error {
	l.Error(err)
	return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error")
}

func NotInService(l logger.Logger) error {
	l.Error("not in service")
	return echo.NewHTTPError(http.StatusInternalServerError, "not in service")
}
