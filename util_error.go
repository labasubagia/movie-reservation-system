package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
)

type Trace struct {
	Function string `json:"function"`
	File     string `json:"file"`
}

type ErrType string

const (
	ErrInternal ErrType = "ErrorInternal"
	ErrInput    ErrType = "ErrorInput"
	ErrNotFound ErrType = "ErrorNotFound"
)

type Err struct {
	Type       ErrType
	Inner      error
	Message    string
	Stacktrace []Trace
}

func (e *Err) Error() string {
	return e.Message
}

func NewErr(typ ErrType, err error, msgFmt string, msgArgs ...any) error {
	return &Err{
		Type:       typ,
		Inner:      err,
		Message:    fmt.Sprintf(msgFmt, msgArgs...),
		Stacktrace: stacktrace(),
	}
}

func ErrIs(err error, typ ErrType) bool {
	if err == nil {
		return false
	}
	var e *Err
	if errors.As(err, &e) {
		return e.Type == typ
	}
	return false
}

func stacktrace() []Trace {
	buf := make([]uintptr, 10)
	n := runtime.Callers(3, buf)
	frames := runtime.CallersFrames(buf[:n])
	var res []Trace
	for {
		frame, more := frames.Next()
		res = append(res, Trace{
			Function: frame.Function,
			File:     fmt.Sprintf("%s:%d", frame.File, frame.Line),
		})
		if !more {
			break
		}
	}
	return res
}

// drivers
func NewSQLErr(err error) error {
	if err == nil {
		return nil
	}

	var p *pgconn.PgError
	if errors.As(err, &p) {
		switch p.Code {
		case pgerrcode.UniqueViolation:
			return NewErr(ErrInput, err, "data already exists!")
		case pgerrcode.ForeignKeyViolation:
			return NewErr(ErrInput, err, "input invalid, check and try again later!")
		}
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return NewErr(ErrNotFound, err, "data not found")
	}

	return NewErr(ErrInternal, err, "something went wrong")
}

// drivers
func NewAPIErr(c echo.Context, err error) error {
	output := c.Get(KeyOutput)
	logInfo := map[string]any{
		"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
		"url":        c.Request().URL.String(),
		"method":     c.Request().Method,
		KeyInput:     c.Get(KeyInput),
		KeyOutput:    output,
	}

	if err == nil {
		return c.JSON(http.StatusOK, Response[any]{
			Message: "ok",
			Data:    output,
		})
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

	return c.JSON(code, Response[any]{
		Message: msg,
		Data:    nil,
	})
}
