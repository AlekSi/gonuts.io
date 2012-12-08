package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"gonuts"
)

func docHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
