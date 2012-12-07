package controllers

import (
	"bytes"
	"html/template"
	"net/http"

	"gonuts"
)

func docHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var content bytes.Buffer
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, "doc.html", ""))

	bd := BaseData{
		Tabtitle: "Documentation",
		Title:    "Documentation",
		Content:  template.HTML(content.String()),
	}
	gonuts.PanicIfErr(Base.Execute(w, &bd))
}
