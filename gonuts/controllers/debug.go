package controllers

import (
	"appengine"
	"appengine/datastore"
	"net/http"

	"gonuts"
)

// FIXME auth
func ahPrepareTestHandler(w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)
	c := appengine.NewContext(r)

	for i := datastore.NewQuery("Version").Filter("Vendor=", "debug").KeysOnly().Run(c); ; {
		key, err := i.Next(nil)
		if err == datastore.Done {
			break
		}
		if err == nil {
			err = datastore.Delete(c, key)
		}
		if err != nil {
			gonuts.LogError(c, err)
			ServeJSONError(w, http.StatusInternalServerError, err, d)
			return
		}
	}

	var nut gonuts.Nut
	for i := datastore.NewQuery("Nut").Filter("Vendor=", "debug").Run(c); ; {
		key, err := i.Next(&nut)
		if err == datastore.Done {
			break
		}
		if err == nil {
			err = datastore.Delete(c, key)
		}
		if err == nil {
			// err = gonuts.RemoveFromSearchIndex(c, &nut)
		}
		if err != nil {
			gonuts.LogError(c, err)
			ServeJSONError(w, http.StatusInternalServerError, err, d)
			return
		}
	}

	d["Message"] = "OK"
	ServeJSON(w, http.StatusOK, d)
	return
}
