package controllers

import (
	"appengine"
	"appengine/datastore"
	gaeuser "appengine/user"
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"gonuts"
	nutp "gonuts.io/AlekSi/nut"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)
	c := appengine.NewContext(r)
	u := gaeuser.Current(c)

	if u == nil || u.ID == "" {
		url, err := gaeuser.LoginURL(c, "/-/me")
		gonuts.LogError(c, err)
		d["LoginURL"] = url
	} else {
		user := new(gonuts.User)
		err := datastore.Get(c, gonuts.UserKey(c, u), user)
		if err == nil {
			url, err := gaeuser.LogoutURL(c, "/")
			gonuts.LogError(c, err)
			d["LogoutURL"] = url
			d["Email"] = user.Email
			d["Token"] = user.Token
			d["GenerateURL"] = "/-/me/generate"
			d["Vendors"] = user.Vendors
		} else if err == datastore.ErrNoSuchEntity {
			url, err := gaeuser.LogoutURL(c, "/-/me")
			gonuts.LogError(c, err)
			d["LogoutURL"] = url
			d["Email"] = u.Email
			d["RegisterURL"] = "/-/me/register"
		} else {
			panic(err)
		}
	}

	var content bytes.Buffer
	gonuts.PanicIfErr(Base.ExecuteTemplate(&content, "me.html", d))

	bd := BaseData{
		Tabtitle: "Me",
		Title:    "Me",
		Content:  template.HTML(content.String()),
	}
	gonuts.PanicIfErr(Base.Execute(w, &bd))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := gaeuser.Current(c)
	if u != nil && u.ID != "" {
		var err error
		vendor := r.FormValue("vendor")
		if !nutp.VendorRegexp.MatchString(vendor) {
			err = fmt.Errorf("Vendor name should match %s.", nutp.VendorRegexp.String())
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		v := &gonuts.Vendor{Vendor: vendor}
		user := &gonuts.User{Id: u.ID, Email: u.Email}
		gonuts.PanicIfErr(user.GenerateToken())

		err = datastore.Get(c, gonuts.VendorKey(c, vendor), v)
		if err != datastore.ErrNoSuchEntity {
			if err == nil {
				err = fmt.Errorf("Vendor name %q is already registered.", vendor)
				WriteError(w, http.StatusForbidden, err)
				return
			}

			gonuts.LogError(c, err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = datastore.Get(c, gonuts.UserKey(c, u), user)
		if err != datastore.ErrNoSuchEntity {
			gonuts.LogError(c, err)
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		c.Infof("Adding user %s (%s) to vendor %s.", user.Id, user.Email, v.Vendor)
		user.AddVendor(v)
		_, err = datastore.Put(c, gonuts.VendorKey(c, vendor), v)
		gonuts.PanicIfErr(err)
		_, err = datastore.Put(c, gonuts.UserKey(c, u), user)
		gonuts.PanicIfErr(err)
	}

	http.Redirect(w, r, "/-/me", 303)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := gaeuser.Current(c)
	if u != nil && u.ID != "" {
		key := gonuts.UserKey(c, u)
		user := gonuts.User{}
		err := datastore.Get(c, key, &user)
		if err == nil {
			gonuts.PanicIfErr(user.GenerateToken())
			_, err = datastore.Put(c, key, &user)
		}
		gonuts.LogError(c, err)
	}
	http.Redirect(w, r, "/-/me", 303)
}
