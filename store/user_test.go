package store

import (
	"testing"
)

func Test_CreateUser(t *testing.T) {
	type CreateUser struct {
		name     string
		password string
		mail     string
	}
	createUsers := []CreateUser{
		{"Emma", "PwEmma", "emma@funny.org"},
		{"John", "PwJohn", "john@funny.org"},
		{"Bill", "PwBill", ""},
	}
	store, err := OpenTestStore("Test_CreateUser")
	if err != nil {
		t.Errorf("Test_CreateUser - cannot open store : %s", err.Error())
		return
	}
	users := make([]*User, len(createUsers))
	IDs := make([]uint64, len(createUsers))
	for i, u := range createUsers {
		users[i], err = store.CreateUser(u.name, u.password, u.mail)
		if err != nil {
			t.Errorf("Test_CreateUser - cannot create user %v\n%s", u.name, err.Error())
			return
		}
		IDs[i] = users[i].ID
	}
	for i, u := range users {
		user := store.LookupUser(IDs[i])
		if user == nil {
			t.Errorf("Test_CreateUser - failed to find User %v", u.ID)
			return
		}
		if user.ID != u.ID || user.Name != u.Name {
			t.Errorf("Test_CreateUser - lookup User %v returns different user", user.ID)
			return
		}
	}
	for _, u := range users {
		user := store.LookupUserByName(u.Name)
		if user == nil {
			t.Errorf("Test_CreateUser - failed to find User by name %v", u.Name)
			return
		}
		if user.ID != u.ID || user.Name != u.Name {
			t.Errorf("Test_CreateUser - lookup User %v returns different user", user.ID)
			return
		}
	}
	for _, u := range users {
		user, err := store.CreateUser(u.Name, "funnyPW", u.TentativeMail.Value)
		if err == nil {
			t.Errorf("Test_CreateUser - created two users with identical name %v: %v and %v", u.Name, u.ID, user.ID)
			return
		}
		if StoreErrorCode(err) != STORE_ERROR_USER_NAME_EXISTS {
			t.Errorf("Test_CreateUser - expected StoreErrorCode %d but got %d", STORE_ERROR_USER_NAME_EXISTS, StoreErrorCode(err))
			return
		}
	}
	if err := store.DeleteUser(users[1].ID); err != nil {
		t.Errorf("Test_CreateUser - Delete user failed %v\n%s", users[1].ID, err.Error())
		return
	}
	users[1] = nil

	for i, u := range createUsers {
		user := store.LookupUser(IDs[i])
		if user == nil && users[i] == nil {
			continue
		}
		if user == nil && users[i] != nil {
			t.Errorf("Test_CreateUser - failed to find User %v after Delete", IDs[i])
			return
		}
		if user != nil && users[i] == nil {
			t.Errorf("Test_CreateUser - found deleted User %v", createUsers[i].name)
			return
		}
		if user != nil && (user.ID != users[i].ID || user.Name != u.name) {
			t.Errorf("Test_CreateUser - lookup User %v after Delete returns different user after Delete", IDs[i])
			return
		}
	}
	for i, u := range createUsers {
		user := store.LookupUserByName(u.name)
		if user == nil && users[i] == nil {
			continue
		}
		if user == nil && users[i] != nil {
			t.Errorf("Test_CreateUser - failed to find User %v after Delete", users[i].ID)
			return
		}
		if user != nil && users[i] == nil {
			t.Errorf("Test_CreateUser - found deleted User %v", users[i].ID)
			return
		}
		if user != nil && (user.ID != users[i].ID || user.Name != u.name) {
			t.Errorf("Test_CreateUser - lookup User %v after Delete returns different user after Delete", user.ID)
			return
		}
	}
}
