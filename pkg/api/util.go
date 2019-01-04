package api

import "fmt"

type C struct {
	Valid bool
	Error error
}

func Cb(chk bool, errmsg string, params ...interface{}) C {
	return C{chk, fmt.Errorf(errmsg, params...)}
}

func Ce(err error) C {
	return C{err == nil, err}
}

func Check(args ...C) error {
	for _, c := range args {
		if !c.Valid {
			return c.Error
		}
	}
	return nil
}
