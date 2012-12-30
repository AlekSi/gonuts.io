package controllers

import (
	"fmt"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"gonuts"
)

func ahHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	d := make(ContentData)

	d["Message"] = "Hello from _ah."
	ServeJSON(w, http.StatusOK, d)
	return
}

func ahCronSearchHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer r.Body.Close()
	d := make(ContentData)
	c := appengine.NewContext(r)

	var nut gonuts.Nut
	for i := datastore.NewQuery("Nut").Run(c); ; {
		_, err := i.Next(&nut)
		if err == datastore.Done {
			break
		}
		if err == nil {
			err = gonuts.AddToSearchIndex(c, &nut)
		}
		if err != nil {
			gonuts.LogError(c, err)
			ServeJSONError(w, http.StatusInternalServerError, err, d)
			return
		}
	}

	d["Message"] = fmt.Sprintf("Cron done in %f seconds.", time.Since(start)/time.Second)
	ServeJSON(w, http.StatusOK, d)
	return
}
