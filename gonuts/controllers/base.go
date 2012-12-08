package controllers

import (
	"encoding/json"
	_ "expvar"
	"html/template"
	"net/http"
	"strings"

	"gonuts"
	"gopath/src/github.com/bmizerany/pat"
)

type ContentData map[string]interface{}

type BaseData struct {
	Tabtitle string
	Title    string
	Subtitle string
	Content  template.HTML
}

func ServeJSON(w http.ResponseWriter, code int, d ContentData) {
	b, err := json.Marshal(d)
	gonuts.PanicIfErr(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(b)
	gonuts.PanicIfErr(err)
}

func ServeJSONError(w http.ResponseWriter, code int, err error, d ContentData) {
	d["Message"] = err.Error()
	ServeJSON(w, code, d)
}

var (
	Router = pat.New()
	Base   = template.Must(template.ParseFiles("gonuts/templates/base.html"))
)

func init() {
	http.Handle("/", Router)

	Router.Get("/_ah/", http.HandlerFunc(ahHandler))

	Router.Get("/-/about", http.HandlerFunc(aboutHandler))
	Router.Get("/-/doc", http.HandlerFunc(docHandler))
	Router.Get("/-/doc/:section", http.HandlerFunc(docHandler))
	Router.Get("/-/me", http.HandlerFunc(myHandler))
	Router.Get("/-/me/register", http.HandlerFunc(registerHandler))
	Router.Get("/-/nuts", http.HandlerFunc(nutsHandler))

	Router.Put("/:name/:version", http.HandlerFunc(nutCreateHandler))
	Router.Get("/:name/:version", http.HandlerFunc(nutShowHandler))
	Router.Get("/:name", http.HandlerFunc(nutShowHandler))

	Router.Get("/", http.HandlerFunc(welcomeHandler))

	Base.Funcs(map[string]interface{}{"lower": strings.ToLower})
	template.Must(Base.ParseGlob("gonuts/templates/base/*.html"))
}
