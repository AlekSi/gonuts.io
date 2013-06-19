package gonuts

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"time"
)

type User struct {
	// StringID (entity name, key name) is appengine/user.User.ID
	Id                string // same as StringID
	Email             string
	FederatedIdentity string
	Token             string
	Debug             bool
	Vendors           []string
}

func UserKey(c appengine.Context, u *user.User) *datastore.Key {
	return datastore.NewKey(c, "User", u.ID, 0, nil)
}

func (user *User) AddVendor(vendor *Vendor) {
	vmap := make(map[string]bool, len(user.Vendors)+1)
	for _, v := range user.Vendors {
		vmap[v] = true
	}
	vmap[vendor.Vendor] = true

	vendors := make([]string, 0, len(vmap))
	for v := range vmap {
		vendors = append(vendors, v)
	}
	sort.Strings(vendors)
	user.Vendors = vendors

	umap := make(map[string]bool, len(vendor.UserStringID)+1)
	for _, u := range vendor.UserStringID {
		umap[u] = true
	}
	umap[user.Id] = true

	users := make([]string, 0, len(umap))
	for u := range umap {
		users = append(users, u)
	}
	sort.Strings(users)
	vendor.UserStringID = users
}

func (user *User) Identifier() (id string) {
	id = user.Email
	if id == "" {
		id = user.FederatedIdentity
	}
	if id == "" {
		panic(fmt.Errorf("User %#v has neither email nor federated identity.", user))
	}
	return
}

type Vendor struct {
	// StringID (entity name, key name) is "Vendor"
	Vendor       string
	UserStringID []string // slice of User.StringID
}

func VendorKey(c appengine.Context, vendor string) *datastore.Key {
	return datastore.NewKey(c, "Vendor", vendor, 0, nil)
}

type Nut struct {
	// StringID (entity name, key name) is "Vendor/Name"
	Vendor string
	Name   string
	Doc    string // Doc of latest published version
	// TODO store total number of downloads, update by cron
}

func NutKey(c appengine.Context, vendor, nut string) *datastore.Key {
	return datastore.NewKey(c, "Nut", fmt.Sprintf("%s/%s", vendor, nut), 0, nil)
}

type Version struct {
	// StringID (entity name, key name) is "Vendor/Name-Version"
	Vendor     string
	Name       string
	Version    string
	VersionNum int // for sorting
	Doc        string
	Homepage   string
	BlobKey    appengine.BlobKey
	CreatedAt  time.Time
	Downloads  int
}

func VersionKey(c appengine.Context, vendor, nut, version string) *datastore.Key {
	return datastore.NewKey(c, "Version", fmt.Sprintf("%s/%s-%s", vendor, nut, version), 0, nil)
}

func (user *User) GenerateToken() (err error) {
	const n = 4096

	buf := make([]byte, n)
	_, err = io.ReadAtLeast(rand.Reader, buf, n)
	if err != nil {
		return
	}

	h := sha1.New()
	_, err = h.Write(buf)
	if err != nil {
		return
	}

	user.Token = hex.EncodeToString(h.Sum(nil))
	return
}
