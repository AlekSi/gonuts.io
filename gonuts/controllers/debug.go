package controllers

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"net/http"

	"gonuts"
)

func debugPrepareTestHandler(w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)
	c := appengine.NewContext(r)

	// extract token from request
	token := r.URL.Query().Get("token")
	if token == "" {
		ServeJSONError(w, http.StatusForbidden, fmt.Errorf("Can't find 'token' in get parameters."), d)
		return
	}

	// find user by token
	q := datastore.NewQuery("User").Filter("Token=", token)
	var users []gonuts.User
	_, err := q.Limit(2).GetAll(c, &users)
	if err != nil || len(users) != 1 {
		if err == nil || err == datastore.ErrNoSuchEntity {
			err = fmt.Errorf("Can't find user with token %q.", token)
		}
		ServeJSONError(w, http.StatusForbidden, err, d)
		return
	}

	// user should be debug
	if !users[0].Debug {
		err = fmt.Errorf("User doesn't have debug access.")
		ServeJSONError(w, http.StatusForbidden, err, d)
		return
	}

	// remove versions
	for i := datastore.NewQuery("Version").Filter("Vendor=", "debug").KeysOnly().Run(c); ; {
		key, err := i.Next(nil)
		if err == datastore.Done {
			break
		}
		if err == nil {
			err = datastore.Delete(c, key)
		}
		if err != nil {
			gonuts.LogError(c, err)
			ServeJSONError(w, http.StatusInternalServerError, err, d)
			return
		}
	}

	// remove nuts
	var nut gonuts.Nut
	for i := datastore.NewQuery("Nut").Filter("Vendor=", "debug").Run(c); ; {
		key, err := i.Next(&nut)
		if err == datastore.Done {
			break
		}
		if err == nil {
			err = datastore.Delete(c, key)
		}
		if err == nil {
			err = gonuts.RemoveFromSearchIndex(c, &nut)
		}
		if err != nil {
			gonuts.LogError(c, err)
			ServeJSONError(w, http.StatusInternalServerError, err, d)
			return
		}
	}

	d["Message"] = "OK"
	ServeJSON(w, http.StatusOK, d)
	return
}
