package controllers

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"bytes"
	"html/template"
	"net/http"

	"gonuts"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	d := make(ContentData)
	c := appengine.NewContext(r)
	u := user.Current(c)

	if u == nil || u.ID == "" {
		url, err := user.LoginURL(c, "/-/me")
		gonuts.LogError(c, err)
		d["LoginURL"] = url
	} else {
		d["RegisterURL"] = "/-/me/register"

		gnu := new(gonuts.User)
		err := datastore.Get(c, datastore.NewKey(c, "User", u.ID, 0, nil), gnu)
		if err == nil {
			url, err := user.LogoutURL(c, "/")
			gonuts.LogError(c, err)
			d["LogoutURL"] = url
			d["Email"] = gnu.Email
			d["Token"] = gnu.Token
		} else if err == datastore.ErrNoSuchEntity {
			url, err := user.LogoutURL(c, "/-/me")
			gonuts.LogError(c, err)
			d["Email"] = u.Email
			d["LogoutURL"] = url
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
	u := user.Current(c)
	if u == nil || u.ID == "" {
		http.Redirect(w, r, "/-/me", 303)
	}

	gnu := gonuts.User{Email: u.Email}
	gonuts.PanicIfErr(gnu.GenerateToken())
	key := datastore.NewKey(c, "User", u.ID, 0, nil)
	_, err := datastore.Put(c, key, &gnu)
	gonuts.PanicIfErr(err)
	http.Redirect(w, r, "/-/me", 303)
}
