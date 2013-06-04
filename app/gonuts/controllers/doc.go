package controllers

import (
	"appengine"
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"gonuts"
)

func docHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var content bytes.Buffer

	section := r.URL.Query().Get(":section")
	file := "doc.html"
	if section != "" {
		file = fmt.Sprintf("doc_%s.html", section)
	}
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, file, ""))

	bd := BaseData{
		Tabtitle: "Documentation",
		Title:    "Documentation",
		Content:  template.HTML(content.String()),
	}
	gonuts.PanicIfErr(Base.Execute(w, &bd))
}
