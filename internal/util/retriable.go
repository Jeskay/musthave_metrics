package util

import (
	"errors"
	"syscall"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var tries = []time.Duration{
	time.Second,
	time.Second * 3,
	time.Second * 5,
}

func TryRun(f func() error, condition func(error) bool) (err error) {
	return tryRunRec(f, condition, 0)
}

func IsPGConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
}

func IsConnectionRefused(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED)
}

func tryRunRec(f func() error, condition func(error) bool, try int) error {
	err := f()
	if condition(err) {
		if try >= len(tries) {
			return err
		}
		time.Sleep(tries[try])
		return tryRunRec(f, condition, try+1)
	}
	return err
}
