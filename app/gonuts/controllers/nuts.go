package controllers

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"gonuts"
)

func nutsHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)
	apiCall := r.Header.Get("Accept") == "application/json"

	// TODO: no need to load all, then render all - replace with chunking
	var nuts []gonuts.Nut
	var err error
	var title string
	vendor := r.URL.Query().Get(":vendor")
	q := r.URL.Query().Get("q")
	if vendor != "" {
		title = fmt.Sprintf("%s's Nuts", vendor)
		_, err = datastore.NewQuery("Nut").Filter("Vendor=", vendor).Order("Name").GetAll(c, &nuts)
	} else if q == "" {
		title = "All Nuts"
		_, err = datastore.NewQuery("Nut").Order("Vendor").Order("Name").GetAll(c, &nuts)
	} else {
		title = fmt.Sprintf("Search %q", q)
		res, err := gonuts.SearchIndex(c, q)
		gonuts.LogError(c, err)
		keys := make([]*datastore.Key, len(res))
		for i, pair := range res {
			keys[i] = gonuts.NutKey(c, pair[0], pair[1])
		}
		nuts = make([]gonuts.Nut, len(keys))
		err = datastore.GetMulti(c, keys, nuts)
	}
	gonuts.LogError(c, err)
	d["Nuts"] = nuts
	status := http.StatusOK
	if len(nuts) == 0 {
		status = http.StatusNotFound
	}

	if apiCall {
		d["Message"] = title
		ServeJSON(w, status, d)
		return
	}

	var content bytes.Buffer
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, "nuts.html", d))

	bd := BaseData{
		Tabtitle: title,
		Title:    title,
		Content:  template.HTML(content.String()),
	}

	w.WriteHeader(status)
	gonuts.PanicIfErr(Base.Execute(w, &bd))
}
