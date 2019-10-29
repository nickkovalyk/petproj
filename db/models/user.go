package models

import (
	"encoding/json"
)

type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	FirstName  string `json:"firstName" db:"first_name"`
	LastName   string `json:"lastName" db:"last_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Phone      string `json:"phone"`
	UserStatus int    `json:"userStatus" db:"user_status"`
}

func (u *User) MarshalJSON() (output []byte, err error) {
	type Alias User
	u.Password = ""
	return json.Marshal((*Alias)(u))
}

func (u *User) Validate() error {
	if len(u.Username) < 6 {
		return ValidationError("Username must not be less than 6 characters")
	}
	if len(u.Password) < 6 {
		return ValidationError("Password must not be less than 6 characters")

	}
	return nil
}
