package gonuts

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"io"

	"appengine"
)

type User struct {
	// StringID (entity name, key name) is appengine/user.User.ID
	Email string
	Token string
}

// There StringID() equals nut name for fast gets, and Name equals nut name for sorting.
type Nut struct {
	// StringID (entity name, key name) is nut name
	Name         string
	Doc          string   // Doc of latest published version
	UserStringID []string // slice of User.StringID
}

// There StringID() contains nut name for fast gets, and NutName equals nut name for queries.
type Version struct {
	// StringID (entity name, key name) is "NutName-Version"
	NutName    string
	Version    string
	VersionNum int
	Doc        string
	BlobKey    appengine.BlobKey
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
