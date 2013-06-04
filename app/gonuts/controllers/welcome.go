package controllers

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"html/template"
	"net/http"

	"gonuts"
)

func welcomeHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)

	nuts, err := datastore.NewQuery("Version").Count(c)
	gonuts.LogError(c, err)
	d["VersionCount"] = nuts

	nuts, err = datastore.NewQuery("Nut").Count(c)
	gonuts.LogError(c, err)
	d["NutsCount"] = nuts

	users, err := datastore.NewQuery("User").Count(c)
	gonuts.LogError(c, err)
	d["UsersCount"] = users

	var content bytes.Buffer
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, "welcome.html", d))

	bd := BaseData{
		Content: template.HTML(content.String()),
	}
	gonuts.PanicIfErr(Base.Execute(w, &bd))
}
