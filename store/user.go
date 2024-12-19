package store

import (
	"fmt"
	"strings"
	"time"

	bh "github.com/timshannon/bolthold"
)

type Tentative struct {
	Value      string
	ValueHash  string
	Token      string
	ExpiryTime time.Time
}

type User struct {
	ID                uint64 `boltholdKey:"ID"`
	Name              string `boltholdUnique:"UniqueName"`
	PasswordHash      string
	CreationTime      time.Time
	Token             string
	MailHash          string
	TentativeMail     Tentative
	TentativePassword Tentative
}

const (
	minNameLen     = 4
	minPasswordLen = 6
)

func (store *_Store) CreateUser(name string, password string, mail string) (*User, error) {
	if len(name) < minNameLen {
		return nil, NewStoreError(STORE_ERROR_USER_NAME_SHORT, fmt.Sprintf("user name must contain at least %d characters", minNameLen))
	}
	if len(password) < minPasswordLen {
		return nil, NewStoreError(STORE_ERROR_USER_NAME_SHORT, fmt.Sprintf("user password must contain at least %d characters", minPasswordLen))
	}
	passwordHash := CryptoHash(password)
	mailHash := CryptoHash(mail)
	user := &User{
		ID:           0,
		Name:         name,
		PasswordHash: passwordHash,
		Token:        CreateToken(),
		CreationTime: time.Now(),
	}
	if len(mail) > 0 {
		user.TentativeMail = Tentative{
			Value:      mail,
			ValueHash:  mailHash,
			Token:      CreateToken(),
			ExpiryTime: time.Now().Add(time.Hour * 12),
		}
	}
	if err := store.db.Insert(bh.NextSequence(), user); err != nil {
		if err == bh.ErrUniqueExists {
			return nil, NewStoreError(STORE_ERROR_USER_NAME_EXISTS, fmt.Sprintf(`user name "%s" allready exists`, user.Name))
		}
		return nil, NewStoreError(STORE_ERROR_UNEXPECTED, err.Error())
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
