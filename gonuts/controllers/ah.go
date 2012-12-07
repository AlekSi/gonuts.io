package controllers

import (
	"net/http"

	"appengine"
)

func ahHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	c := appengine.NewContext(r)
	c.Debugf("Request to %s", r.URL)
	w.WriteHeader(http.StatusOK)
}
