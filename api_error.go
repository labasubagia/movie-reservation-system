package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HTTPErrorHandler(err error, c echo.Context) {
	output := c.Get(KeyOutput)
	logInfo := map[string]any{
		"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
		"url":        c.Request().URL.String(),
		"method":     c.Request().Method,
		KeyInput:     c.Get(KeyInput),
		KeyOutput:    output,
	}

	if err == nil {
		c.JSON(http.StatusOK, Response{
			Message: "ok",
			Data:    output,
		})
		return
	}

	code := http.StatusInternalServerError
	msg := ""
	var e *Err
	if errors.As(err, &e) {
		msg = e.Message
		switch e.Type {
		case ErrInput:
			code = http.StatusBadRequest
		case ErrNotFound:
			code = http.StatusNotFound
		case ErrInternal:
			code = http.StatusInternalServerError
			logInfo["stacktrace"] = e.Stacktrace
			if e.Inner != nil {
				logInfo["error"] = e.Inner.Error()
			}
		}
	}

	var h *echo.HTTPError
	if errors.As(err, &h) {
		code = h.Code
		msg = h.Message.(string)
	}

	if code == http.StatusInternalServerError {
		logInfo["status_code"] = code
		logInfo["message"] = msg
		if _, ok := logInfo["error"]; !ok {
			logInfo["error"] = err.Error()
		}
		c.Logger().Errorj(logInfo)
	}

	c.JSON(code, Response{
		Message: msg,
		Data:    nil,
	})
}
