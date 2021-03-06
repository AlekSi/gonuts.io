package controllers

import (
	"appengine"
	"bytes"
	"html/template"
	"net/http"

	"gonuts"
)

func aboutHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var content bytes.Buffer
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, "about.html", ""))

	bd := BaseData{
		Tabtitle: "About",
		Title:    "About",
		Content:  template.HTML(content.String()),
	}
	gonuts.PanicIfErr(Base.Execute(w, &bd))
}
