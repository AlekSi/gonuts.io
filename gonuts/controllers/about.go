package controllers

import (
	"bytes"
	"html/template"
	"net/http"

	"gonuts"
)

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var content bytes.Buffer
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, "about.html", ""))

	bd := BaseData{
		Tabtitle: "About",
		Title:    "About",
		Content:  template.HTML(content.String()),
	}
	gonuts.PanicIfErr(Base.Execute(w, &bd))
}
