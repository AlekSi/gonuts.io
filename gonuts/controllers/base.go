package controllers

import (
	"encoding/json"
	_ "expvar"
	"fmt"
	"html/template"
	"net/http"
	_ "net/http/pprof"

	"github.com/bmizerany/pat"
	"gonuts"
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

func WriteError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	_, err = w.Write([]byte(fmt.Sprintf("%s", err)))
	gonuts.PanicIfErr(err)
}

var (
	Router = pat.New()
	Base   *template.Template
)

func init() {
	Router.Get("/_ah/cron/search", http.HandlerFunc(ahCronSearchHandler))
	Router.Get("/_ah/ad-hoc", http.HandlerFunc(ahAdHoc))
	Router.Get("/_ah/", http.HandlerFunc(ahHandler))

	Router.Get("/debug/prepare_test", http.HandlerFunc(debugPrepareTestHandler))

	Router.Get("/-/about", http.HandlerFunc(aboutHandler))
	Router.Get("/-/doc", http.HandlerFunc(docHandler))
	Router.Get("/-/doc/:section", http.HandlerFunc(docHandler))
	Router.Get("/-/me", http.HandlerFunc(myHandler))
	Router.Post("/-/me/register", http.HandlerFunc(registerHandler))
	Router.Get("/-/me/generate", http.HandlerFunc(generateHandler))
	Router.Get("/-/nuts", http.HandlerFunc(nutsHandler))

	Router.Put("/:vendor/:name/:version", http.HandlerFunc(nutCreateHandler))
	Router.Get("/:vendor/:name/:version", http.HandlerFunc(nutShowHandler))
	Router.Get("/:vendor/:name", http.HandlerFunc(nutShowHandler))
	Router.Get("/:vendor", http.HandlerFunc(nutsHandler))

	Router.Get("/", http.HandlerFunc(welcomeHandler))

	http.Handle("/", Router)

	Base = template.Must(template.ParseFiles("gonuts/templates/base.html"))
	template.Must(Base.ParseGlob("gonuts/templates/base/*.html"))
}
