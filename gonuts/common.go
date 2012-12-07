package gonuts

import (
	"appengine"
)

func LogError(c appengine.Context, err error) {
	if err != nil {
		c.Warningf("%s", err)
	}
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
