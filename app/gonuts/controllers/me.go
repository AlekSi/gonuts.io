package controllers

import (
	"appengine"
	"appengine/datastore"
	gaeuser "appengine/user"
	"bytes"
	"fmt"
	"gonuts"
	nutp "gonuts.io/AlekSi/nut"
	"html/template"
	"net/http"
	"net/url"
)

func myHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)
	u := gaeuser.Current(c)

	if u == nil || u.ID == "" {
		url, err := gaeuser.LoginURL(c, "/-/me")
		gonuts.LogError(c, err)
		d["LoginURL"] = url
		d["OpenIDURL"] = "/-/me/openid"
	} else {
		user := new(gonuts.User)
		err := datastore.Get(c, gonuts.UserKey(c, u), user)
		if err == nil {
			url, err := gaeuser.LogoutURL(c, "/")
			gonuts.LogError(c, err)
			d["LogoutURL"] = url
			d["Identifier"], _ = user.Identifier()
			d["Token"] = user.Token
			d["GenerateURL"] = "/-/me/generate"
			d["Vendors"] = user.Vendors
		} else if err == datastore.ErrNoSuchEntity {
			user = &gonuts.User{Id: u.ID, Email: u.Email, FederatedIdentity: u.FederatedIdentity}
			url, err := gaeuser.LogoutURL(c, "/-/me")
			gonuts.LogError(c, err)
			d["LogoutURL"] = url
			d["Identifier"], _ = user.Identifier()
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

func registerHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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
		user := &gonuts.User{Id: u.ID, Email: u.Email, FederatedIdentity: u.FederatedIdentity}
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

		id, _ := user.Identifier()
		c.Infof("Adding user %s (%s) to vendor %s.", user.Id, id, v.Vendor)
		user.AddVendor(v)
		_, err = datastore.Put(c, gonuts.VendorKey(c, vendor), v)
		gonuts.PanicIfErr(err)
		_, err = datastore.Put(c, gonuts.UserKey(c, u), user)
		gonuts.PanicIfErr(err)
	}

	http.Redirect(w, r, "/-/me", 303)
}

func generateHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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

func openIdHandler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := gaeuser.Current(c)
	if u != nil && u.ID != "" {
		err := fmt.Errorf("This page is not supposed to be accessible by logged-in users")
		WriteError(w, http.StatusForbidden, err)
		return
	}

	provider := r.FormValue("provider")
	if _, err := url.Parse(provider); err != nil {
		err := fmt.Errorf("OpenID provider name should be a valid url")
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	url, err := gaeuser.LoginURLFederated(c, "/-/me", provider)
	if err != nil {
		gonuts.LogError(c, err)
		return
	}

	http.Redirect(w, r, url, 303)
}
