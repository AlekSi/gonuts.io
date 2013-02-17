package gonuts

import (
	"appengine"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
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
