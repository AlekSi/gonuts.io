package gonuts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"appengine"
	"appengine/urlfetch"
)

var (
	searchFindUrl url.URL
	searchAddUrl  url.URL
)

func init() {
	searchFindUrl = url.URL{Scheme: "http"}
	if appengine.IsDevAppServer() {
		searchFindUrl.Host = "localhost:8081"
	} else {
		searchFindUrl.Host = "search-gonuts-io.appspot.com"
	}

	searchAddUrl = searchFindUrl
	searchAddUrl.Path = "/add"
	searchAddUrl.RawQuery = fmt.Sprintf("token=%s", addSecretToken)
}

func LogError(c appengine.Context, err error) {
	if err != nil {
		c.Warningf("%s", err)
	}
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func AddToSearchIndex(c appengine.Context, nut *Nut) (err error) {
	m := make(map[string]interface{})
	m["Nut"] = nut
	b, err := json.Marshal(m)
	if err != nil {
		return
	}

	client := urlfetch.Client(c)
	res, err := client.Post(searchAddUrl.String(), "application/json", bytes.NewReader(b))
	if err == nil && res.StatusCode != 201 {
		err = fmt.Errorf("%s -> %d", searchAddUrl.String(), res.StatusCode)
	}
	return
}
