package main

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
