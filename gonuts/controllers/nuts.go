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

func nutsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	d := make(ContentData)
	c := appengine.NewContext(r)
	apiCall := r.Header.Get("Accept") == "application/json"

	// TODO: no need to load all, then render all - replace with chunking
	var nuts []gonuts.Nut
	var err error
	title := "All Nuts"
	q := r.URL.Query().Get("q")
	if q == "" {
		_, err = datastore.NewQuery("Nut").Order("Vendor").Order("Name").GetAll(c, &nuts)
	} else {
		title = fmt.Sprintf("Search %q", q)
		names, err := gonuts.SearchIndex(c, q)
		gonuts.LogError(c, err)

		// FIXME
		keys := make([]*datastore.Key, len(names))
		for i, name := range names {
			keys[i] = datastore.NewKey(c, "Nut", name, 0, nil)
		}
		nuts = make([]gonuts.Nut, len(keys))
		err = datastore.GetMulti(c, keys, nuts)
	}
	gonuts.LogError(c, err)
	d["Nuts"] = nuts

	if apiCall {
		d["Message"] = title
		ServeJSON(w, http.StatusOK, d)
		return
	}

	var content bytes.Buffer
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, "nuts.html", d))

	bd := BaseData{
		Tabtitle: title,
		Title:    title,
		Content:  template.HTML(content.String()),
	}

	gonuts.PanicIfErr(Base.Execute(w, &bd))
}
