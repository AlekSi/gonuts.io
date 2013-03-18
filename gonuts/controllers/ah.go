package controllers

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"net/http"
	"time"

	"gonuts"
)

func ahHandler(w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)

	d["Message"] = "Hello from _ah."
	ServeJSON(w, http.StatusOK, d)
	return
}

func ahAdHoc(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	user := new(gonuts.User)
	gonuts.PanicIfErr(datastore.Get(c, gonuts.UserKey(c, u), user))

	vendor := &gonuts.Vendor{Vendor: "debug"}
	user.AddVendor(vendor)
	_, err := datastore.Put(c, gonuts.VendorKey(c, "debug"), vendor)
	gonuts.PanicIfErr(err)
	_, err = datastore.Put(c, gonuts.UserKey(c, u), user)
	gonuts.PanicIfErr(err)

	w.Write([]byte(fmt.Sprintf("%#v\n", user)))
	w.Write([]byte(fmt.Sprintf("%#v\n", vendor)))
	return
}

func ahCronSearchHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
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

	m := fmt.Sprintf("Search index updated in %d seconds.", time.Since(start)/time.Second)
	c.Infof("%s", m)
	d["Message"] = m
	ServeJSON(w, http.StatusOK, d)
	return
}
