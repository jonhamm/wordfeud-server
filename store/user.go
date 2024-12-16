package store

import (
	"strings"
	"time"

	bh "github.com/timshannon/bolthold"
)

type Tentative struct {
	Value      string
	Token      string
	ExpiryTime time.Time
}

type User struct {
	ID             uint64 `boltholdKey:"ID"`
	Name           string `boltholdUnique:"UniqueName"`
	CreationTime   time.Time
	Token          string
	MailHash       string
	TentativeMail  Tentative
	TentativeToken Tentative
}

func (store *_Store) CreateUser(name string, mail string) (*User, error) {
	user := &User{
		ID:           0,
		Name:         name,
		CreationTime: time.Now(),
		TentativeMail: Tentative{
			Value:      mail,
			Token:      CreateToken(),
			ExpiryTime: time.Now().Add(time.Hour * 12),
		},
	}
	if err := store.db.Insert(bh.NextSequence(), user); err != nil {
		return nil, err
	}
	return user, nil
}

func (store *_Store) LookupUser(ID uint64) *User {

	user := &User{}
	if err := store.db.Get(ID, user); err != nil {
		return nil
	}
	return user
}
func (store *_Store) LookupUserByName(name string) *User {

	user := &User{}
	if err := store.db.FindOne(user, bh.Where("Name").Eq(name)); err != nil {
		return nil
	}
	return user
}

func (store *_Store) DeleteUser(ID uint64) error {
	return store.db.Delete(ID, User{})
}

func mailHash(mail string) string {
	return strings.ToLower(mail)
}
