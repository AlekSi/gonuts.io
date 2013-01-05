package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"appengine"
	"appengine/blobstore"
	"appengine/datastore"
	"gonuts"
	nutp "gopath/src/gonuts.io/nut/0.2.0"
)

func nutCreateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	d := make(ContentData)
	c := appengine.NewContext(r)

	name := r.URL.Query().Get(":name")
	ver := r.URL.Query().Get(":version")

	if name == "" || (ver != "" && !nutp.VersionRegexp.MatchString(ver)) {
		err := fmt.Errorf("Invalid name %q or version %q.", name, ver)
		ServeJSONError(w, http.StatusBadRequest, err, d)
		return
	}

	// extract token from request
	token := r.URL.Query().Get("token")
	if token == "" {
		ServeJSONError(w, http.StatusForbidden, fmt.Errorf("Can't find 'token' in get parameters."), d)
		return
	}

	// find user by token
	q := datastore.NewQuery("User").KeysOnly().Filter("Token=", token)
	userKeys, err := q.Limit(2).GetAll(c, nil)
	if err != nil || len(userKeys) != 1 {
		if err == nil || err == datastore.ErrNoSuchEntity {
			err = fmt.Errorf("Can't find user with token %q.", token)
		}
		ServeJSONError(w, http.StatusForbidden, err, d)
		return
	}
	userID := userKeys[0].StringID()

	// nut should belong to current user
	nutKey := datastore.NewKey(c, "Nut", name, 0, nil)
	nut := gonuts.Nut{Name: name}
	err = datastore.Get(c, nutKey, &nut)
	if err == datastore.ErrNoSuchEntity {
		nut.UserStringID = []string{userID}
	} else if err != nil {
		ServeJSONError(w, http.StatusInternalServerError, err, d)
		return
	}
	found := false
	for _, id := range nut.UserStringID {
		if id == userID {
			found = true
			break
		}
	}
	if !found {
		ServeJSONError(w, http.StatusForbidden, fmt.Errorf("Nut %q already exists and belongs to another user.", name), d)
		return
	}

	// nut version should not exist
	versionKey := datastore.NewKey(c, "Version", fmt.Sprintf("%s-%s", name, ver), 0, nil)
	version := gonuts.Version{NutName: name, Version: ver, CreatedAt: time.Now()}
	err = datastore.Get(c, versionKey, &version)
	if err != nil && err != datastore.ErrNoSuchEntity {
		ServeJSONError(w, http.StatusInternalServerError, err, d)
		return
	}
	if err == nil {
		ServeJSONError(w, http.StatusConflict, fmt.Errorf("Nut %q version %q already exists.", name, ver), d)
		return
	}

	// read nut from request body
	nf := new(nutp.NutFile)
	b, err := ioutil.ReadAll(r.Body)
	if err == nil {
		_, err = nf.ReadFrom(bytes.NewReader(b))
	}
	if err != nil {
		ServeJSONError(w, http.StatusBadRequest, err, d)
		return
	}
	nut.Doc = nf.Doc
	version.Doc = nf.Doc
	version.VersionNum = nf.Version.Major*1000000 + nf.Version.Minor*1000 + nf.Version.Patch // for sorting

	// check name and version match
	if nf.Name != name || nf.Version.String() != ver {
		err = fmt.Errorf("Nut name %q and version %q from URL don't match found in body: %q %q.",
			name, ver, nf.Name, nf.Version.String())
		ServeJSONError(w, http.StatusBadRequest, err, d)
		return
	}

	// check nut
	errors := nf.Check()
	if nf.Name == "debug" {
		errors = append(errors, "Name 'debug' is reserved.")
	}
	if len(errors) != 0 {
		err = fmt.Errorf("%s", strings.Join(errors, "\n"))
		ServeJSONError(w, http.StatusBadRequest, err, d)
		return
	}

	// TODO wrap writes in transaction

	// store nut blob
	bw, err := blobstore.Create(c, r.Header.Get("Content-Type"))
	if err == nil {
		_, err = bw.Write(b)
		if err == nil {
			err = bw.Close()
		}
	}
	if err != nil {
		ServeJSONError(w, http.StatusInternalServerError, err, d)
		return
	}

	// store nut version
	blobKey, err := bw.Key()
	if err == nil {
		version.BlobKey = blobKey
		_, err = datastore.Put(c, versionKey, &version)
	}
	if err != nil {
		ServeJSONError(w, http.StatusInternalServerError, err, d)
		return
	}

	// store nut with new doc
	_, err = datastore.Put(c, nutKey, &nut)
	if err != nil {
		ServeJSONError(w, http.StatusInternalServerError, err, d)
		return
	}

	// update search index
	go func() {
		err := gonuts.AddToSearchIndex(c, &nut)
		gonuts.LogError(c, err)
	}()

	// done!
	d["Message"] = fmt.Sprintf("Nut %q version %q published.", name, ver)
	ServeJSON(w, http.StatusCreated, d)
	return
}

func nutShowHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	d := make(ContentData)
	c := appengine.NewContext(r)
	apiCall := r.Header.Get("Accept") == "application/zip"

	name := r.URL.Query().Get(":name")
	version := r.URL.Query().Get(":version")

	if name == "" || (version != "" && !nutp.VersionRegexp.MatchString(version)) {
		panic(fmt.Sprint(r.URL.Query()))
	}

	v := new(gonuts.Version)
	var err error
	if version == "" {
		q := datastore.NewQuery("Version").Filter("NutName=", name).Order("-VersionNum").Limit(1)
		_, err = q.Run(c).Next(v)
	} else {
		key := datastore.NewKey(c, "Version", fmt.Sprintf("%s-%s", name, version), 0, nil)
		err = datastore.Get(c, key, v)
	}
	gonuts.LogError(c, err)

	var title string
	if v.BlobKey != "" {
		if apiCall {
			blobstore.Send(w, v.BlobKey)
			go func() {
				v.Downloads++
				key := datastore.NewKey(c, "Version", fmt.Sprintf("%s-%s", v.NutName, v.Version), 0, nil)
				_, err := datastore.Put(c, key, v)
				gonuts.LogError(c, err)
			}()
			return
		}

		d["Name"] = v.NutName
		d["Version"] = v.Version
		d["Doc"] = v.Doc
		title = fmt.Sprintf("%s %s", v.NutName, v.Version)
	} else {
		w.WriteHeader(http.StatusNotFound)
		if apiCall {
			return
		}
		title = fmt.Sprintf("Nut %s %s not found", name, version)
	}
	var content bytes.Buffer
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, "nut.html", d))

	bd := BaseData{
		Tabtitle: title,
		Title:    title,
		Content:  template.HTML(content.String()),
	}

	gonuts.PanicIfErr(Base.Execute(w, &bd))
}
