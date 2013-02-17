package controllers

import (
	"appengine"
	"appengine/blobstore"
	"appengine/datastore"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gonuts"
	nutp "gonuts.io/AlekSi/nut"
)

func nutCreateHandler(w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)
	c := appengine.NewContext(r)
	ct := r.Header.Get("Content-Type")
	putNut := ct == "application/zip"

	if !putNut {
		err := fmt.Errorf(`Unexpected Content-Type %q, should be "application/zip".`, ct)
		ServeJSONError(w, http.StatusNotAcceptable, err, d)
		return
	}

	vendor := r.URL.Query().Get(":vendor")
	name := r.URL.Query().Get(":name")
	ver := r.URL.Query().Get(":version")

	if vendor == "" || name == "" || (ver != "" && !nutp.VersionRegexp.MatchString(ver)) {
		err := fmt.Errorf("Invalid vendor %q, name %q or version %q.", vendor, name, ver)
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
	nutKey := gonuts.NutKey(c, vendor, name)
	nut := gonuts.Nut{Vendor: vendor, Name: name}
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
		ServeJSONError(w, http.StatusForbidden, fmt.Errorf("Nut %s/%s already exists and belongs to another user.", vendor, name), d)
		return
	}

	// nut version should not exist
	versionKey := gonuts.VersionKey(c, vendor, name, ver)
	version := gonuts.Version{Vendor: vendor, NutName: name, Version: ver, CreatedAt: time.Now()}
	err = datastore.Get(c, versionKey, &version)
	if err != nil && err != datastore.ErrNoSuchEntity {
		ServeJSONError(w, http.StatusInternalServerError, err, d)
		return
	}
	if err == nil {
		ServeJSONError(w, http.StatusConflict, fmt.Errorf("Nut %s/%s version %s already exists.", vendor, name, ver), d)
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
	version.Homepage = nf.Homepage
	version.VersionNum = nf.Version.Major*1000000 + nf.Version.Minor*1000 + nf.Version.Patch // for sorting

	// check vendor, name and version match
	if nf.Vendor != vendor || nf.Name != name || nf.Version.String() != ver {
		err = fmt.Errorf("Nut vendor %q, name %q and version %q from URL don't match found in body: %q %q %q.",
			vendor, name, ver, nf.Vendor, nf.Name, nf.Version.String())
		ServeJSONError(w, http.StatusBadRequest, err, d)
		return
	}

	// check nut
	errors := nf.Check()
	if len(errors) != 0 {
		err = fmt.Errorf("%s", strings.Join(errors, "\n"))
		ServeJSONError(w, http.StatusBadRequest, err, d)
		return
	}

	// TODO wrap writes in transaction

	// store nut blob
	bw, err := blobstore.Create(c, ct)
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
	d["Message"] = fmt.Sprintf("Nut %s/%s version %s published.", vendor, name, ver)
	ServeJSON(w, http.StatusCreated, d)
	return
}

func nutShowHandler(w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)
	c := appengine.NewContext(r)
	getNut := r.Header.Get("Accept") == "application/zip"

	vendor := r.URL.Query().Get(":vendor")
	name := r.URL.Query().Get(":name")
	ver := r.URL.Query().Get(":version")

	if vendor == "" || name == "" || (ver != "" && !nutp.VersionRegexp.MatchString(ver)) {
		err := fmt.Errorf("Invalid vendor %q, name %q or version %q.", vendor, name, ver)
		ServeJSONError(w, http.StatusBadRequest, err, d)
		return
	}

	v := new(gonuts.Version)
	var err error
	if ver == "" {
		q := datastore.NewQuery("Version").Filter("Vendor=", vendor).Filter("NutName=", name).Order("-VersionNum").Limit(1)
		_, err = q.Run(c).Next(v)
	} else {
		key := gonuts.VersionKey(c, vendor, name, ver)
		err = datastore.Get(c, key, v)
	}
	gonuts.LogError(c, err)

	var title string
	if v.BlobKey != "" {
		if getNut {
			blobstore.Send(w, v.BlobKey)
			go func() {
				v.Downloads++
				key := gonuts.VersionKey(c, v.Vendor, v.NutName, v.Version)
				_, err := datastore.Put(c, key, v)
				gonuts.LogError(c, err)
			}()
			return
		}

		d["Vendor"] = v.Vendor
		d["Name"] = v.NutName
		d["Version"] = v.Version
		d["Doc"] = v.Doc
		d["Homepage"] = v.Homepage
		title = fmt.Sprintf("%s/%s %s", v.Vendor, v.NutName, v.Version)
	} else {
		w.WriteHeader(http.StatusNotFound)
		if getNut {
			return
		}
		title = fmt.Sprintf("Nut %s/%s version %s not found", vendor, name, ver)
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
