package gonuts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"appengine"
	"appengine/urlfetch"
)

var (
	searchFindUrl   url.URL
	searchAddUrl    url.URL
	searchRemoveUrl url.URL
)

func init() {
	searchFindUrl = url.URL{Scheme: "http", Path: "/"}
	if appengine.IsDevAppServer() {
		searchFindUrl.Host = "localhost:8081"
	} else {
		searchFindUrl.Host = "search-gonuts-io.appspot.com"
	}

	searchAddUrl = searchFindUrl
	searchAddUrl.Path = "/add"
	searchAddUrl.RawQuery = fmt.Sprintf("token=%s", addSecretToken)

	searchRemoveUrl = searchFindUrl
	searchRemoveUrl.Path = "/remove"
	searchRemoveUrl.RawQuery = fmt.Sprintf("token=%s", addSecretToken)
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

	if err == nil {
		res.Body.Close()
		if res.StatusCode != 201 {
			err = fmt.Errorf("%s -> %d", searchAddUrl.String(), res.StatusCode)
		}
	}
	return
}

func RemoveFromSearchIndex(c appengine.Context, nutName string) (err error) {
	u := searchRemoveUrl
	values := u.Query()
	values.Set("nut_name", nutName)
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return
	}

	client := urlfetch.Client(c)
	res, err := client.Do(req)

	if err == nil {
		res.Body.Close()
		if res.StatusCode != 204 {
			err = fmt.Errorf("%s -> %d", u.String(), res.StatusCode)
		}
	}
	return
}

func SearchIndex(c appengine.Context, q string) (names []string, err error) {
	client := urlfetch.Client(c)
	u := searchFindUrl
	u.RawQuery = url.Values{"q": []string{q}}.Encode()
	res, err := client.Get(u.String())
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode != 200 {
			err = fmt.Errorf("%s -> %d", u.String(), res.StatusCode)
			return
		}
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}

	nuts := m["Nuts"].([]interface{})
	names = make([]string, len(nuts))
	for i, n := range nuts {
		names[i] = n.(map[string]interface{})["Name"].(string)
	}
	return
}
