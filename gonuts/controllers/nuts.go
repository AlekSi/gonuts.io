package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
	"gonuts"
)

func nutsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	d := make(ContentData)
	c := appengine.NewContext(r)

	// TODO: no need to load all, then render all - replace with chunking
	var nuts []gonuts.Nut
	var err error
	title := "All Nuts"
	q := r.URL.Query().Get("q")
	if q == "" {
		_, err = datastore.NewQuery("Nut").Order("Name").GetAll(c, &nuts)
	} else {
		title = fmt.Sprintf("Search %q (doesn't really work yet)", q)
		_, err = datastore.NewQuery("Nut").Filter("Name=", q).Limit(1).GetAll(c, &nuts)
	}
	gonuts.LogError(c, err)
	d["Nuts"] = nuts

	var content bytes.Buffer
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, "nuts.html", d))

	bd := BaseData{
		Tabtitle: title,
		Title:    title,
		Content:  template.HTML(content.String()),
	}

	gonuts.PanicIfErr(Base.Execute(w, &bd))
}
