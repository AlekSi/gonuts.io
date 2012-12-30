package controllers

import (
	"net/http"
)

func ahHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	d := map[string]interface{}{"Message": "Hello from _ah."}
	ServeJSON(w, http.StatusOK, d)
	return
}

func ahCronSearchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	d := map[string]interface{}{"Message": "Hello from cron."}
	ServeJSON(w, http.StatusOK, d)
	return
}
