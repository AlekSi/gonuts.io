package gonuts

import (
	"appengine"
	"appengine/datastore"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

type User struct {
	// StringID (entity name, key name) is appengine/user.User.ID
	Email string
	Token string
}

// There StringID() equals nut name for fast gets, and Name equals nut name for sorting.
type Nut struct {
	// StringID (entity name, key name) is "Vendor/Name"
	Vendor       string
	Name         string
	Doc          string   // Doc of latest published version
	UserStringID []string // slice of User.StringID
	// TODO store total number of downloads, update by cron
}

func NutKey(c appengine.Context, vendor, name string) *datastore.Key {
	return datastore.NewKey(c, "Nut", fmt.Sprintf("%s/%s", vendor, name), 0, nil)
}

// There StringID() contains nut name for fast gets, and NutName equals nut name for queries.
type Version struct {
	// StringID (entity name, key name) is "Vendor/NutName-Version"
	Vendor     string
	NutName    string
	Version    string
	VersionNum int // for sorting
	Doc        string
	Homepage   string
	BlobKey    appengine.BlobKey
	CreatedAt  time.Time
	Downloads  int
}

func VersionKey(c appengine.Context, vendor, name, version string) *datastore.Key {
	return datastore.NewKey(c, "Version", fmt.Sprintf("%s/%s-%s", vendor, name, version), 0, nil)
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
