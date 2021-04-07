package main

import (
	"errors"
	"fmt"
	"log"
)

// User Model
type User struct {
	ID       int
	Name     string
	Address  string
	MyNumber string
	Votes    int
}

var _userByIdentifier = map[string]User{}

func getUserIdentifier(name, address, mynumber string) string {
	return fmt.Sprintf("%s-%s-%s", name, address, mynumber)
}
func initUsers() {
	rows, err := db.Query("SELECT id, name, address, mynumber, votes FROM users")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		user := User{}
		err = rows.Scan(&user.ID, &user.Name, &user.Address, &user.MyNumber, &user.Votes)
		if err != nil {
			panic(err.Error())
		}
		_userByIdentifier[getUserIdentifier(user.Name, user.Address, user.MyNumber)] = user
	}
	log.Println("load user done")
}

func getUser(name string, address string, myNumber string) (User, error) {
	emptyU := User{}
	user, ok := _userByIdentifier[getUserIdentifier(name, address, myNumber)]
	if !ok {
		return emptyU, errors.New("user not found")
	}
	return user, nil
}
