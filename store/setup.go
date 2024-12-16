package store

import (
	base64 "encoding/base64"
	"encoding/binary"
	"math/rand"
	"os"
	"path"

	bh "github.com/timshannon/bolthold"
)

type Store interface {
	Defer()
	CreateUser(string, string) (*User, error)
	DeleteUser(uint64) error
	LookupUser(uint64) *User
	LookupUserByName(string) *User
}

type _Store struct {
	db       *bh.Store
	filename string
	test     bool
}

func OpenStore(name string) (Store, error) {
	return OpenStoreFile(".store", name, false)

}

func OpenTestStore(name string) (Store, error) {
	return OpenStoreFile("tmp/.store", name, true)

}

func OpenStoreFile(directory string, filename string, test bool) (*_Store, error) {
	var bolthold *bh.Store
	var err error
	if err = os.MkdirAll(directory, 0777); err != nil {
		return nil, err
	}

	filename = path.Join(directory, filename)
	if test {
		os.Remove(filename)
	}

	if bolthold, err = bh.Open(filename, 0600, nil); err != nil {
		return nil, err
	}

	return &_Store{
		db:       bolthold,
		filename: filename,
		test:     false,
	}, nil
}

func (store *_Store) Defer() {
	store.db.Close()
	if store.test {
		os.Remove(store.filename)
	}
}

func CreateToken() string {
	data := Uint64ToBytes(rand.Uint64())
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return string(dst)
}

// Uint64ToBytes converts the given uint64 value to slice of bytes.
func Uint64ToBytes(val uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, val)
	return b
}
