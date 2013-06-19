package controllers

import (
	"encoding/json"
	_ "expvar"
	"fmt"
	"html/template"
	"net/http"
	_ "net/http/pprof"

	"github.com/bmizerany/pat"
	"github.com/mjibson/appstats"
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
	Router.Get("/_ah/cron/search", appstats.NewHandler(ahCronSearchHandler))
	Router.Get("/_ah/status", appstats.NewHandler(ahStatusHandler))
	Router.Get("/_ah/", appstats.NewHandler(ahHandler))

	Router.Get("/debug/prepare_test", appstats.NewHandler(debugPrepareTestHandler))

	Router.Get("/-/about", appstats.NewHandler(aboutHandler))
	Router.Get("/-/doc", appstats.NewHandler(docHandler))
	Router.Get("/-/doc/:section", appstats.NewHandler(docHandler))
	Router.Get("/-/me", appstats.NewHandler(myHandler))
	Router.Post("/-/me/register", appstats.NewHandler(registerHandler))
	Router.Get("/-/me/generate", appstats.NewHandler(generateHandler))
	Router.Post("/-/me/openid", appstats.NewHandler(openIdHandler))
	Router.Get("/-/nuts", appstats.NewHandler(nutsHandler))

	Router.Put("/:vendor/:name/:version", appstats.NewHandler(nutCreateHandler))
	Router.Get("/:vendor/:name/:version", appstats.NewHandler(nutShowHandler))
	Router.Get("/:vendor/:name", appstats.NewHandler(nutShowHandler))
	Router.Get("/:vendor", appstats.NewHandler(nutsHandler))

	Router.Get("/", appstats.NewHandler(welcomeHandler))

	http.Handle("/", Router)

	Base = template.Must(template.ParseFiles("gonuts/templates/base.html"))
	template.Must(Base.ParseGlob("gonuts/templates/base/*.html"))
}
