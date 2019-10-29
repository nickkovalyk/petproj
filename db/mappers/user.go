package mappers

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
)

type UserMapperInterface interface {
	FindByUsername(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Create(*models.User) error
	UpdateByUsername(u *models.User, username string) error
	CreateMany(users []models.User) error
	DeleteByUsername(username string) error
}

type UserMapper struct {
	DB *sqlx.DB
}

func (m UserMapper) FindByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := m.DB.Get(user, "SELECT * FROM users where username=$1", username)
	if err != nil {
		logrus.Error(err)
		if err == sql.ErrNoRows {
			return nil, NotFoundError("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (m UserMapper) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := m.DB.Get(user, "SELECT * FROM users where email=$1", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NotFoundError("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (m UserMapper) Create(u *models.User) error {
	stmt := `INSERT INTO users (username, first_name, last_name, email, password, phone, user_status )
             VALUES (:username, :first_name, :last_name, :email, :password, :phone, :user_status) RETURNING id;`
	var userID int
	params := map[string]interface{}{
		"username":    u.Username,
		"first_name":  u.FirstName,
		"last_name":   u.LastName,
		"email":       u.Email,
		"password":    u.Password,
		"phone":       u.Phone,
		"user_status": u.UserStatus,
	}
	rows, err := m.DB.NamedQuery(stmt, params)
	if err != nil {
		return errors.Wrap(err, "insert user error")
	}
	for rows.Next() {
		err := rows.Scan(&userID)
		if err != nil {
			return errors.Wrap(err, "scan user id error")
		}
	}
	u.ID = userID

	return nil
}

func (m UserMapper) UpdateByUsername(u *models.User, username string) error {
	stmt := `UPDATE users SET username=:username, first_name=:first_name, last_name=:last_name, email=:email, 
                              password=:password, phone=:phone, user_status=:user_status WHERE username=:username`
	params := map[string]interface{}{
		"username":    u.Username,
		"first_name":  u.FirstName,
		"last_name":   u.LastName,
		"email":       u.Email,
		"password":    u.Password,
		"phone":       u.Phone,
		"user_status": u.UserStatus,
	}
	_, err := m.DB.NamedExec(stmt, params)
	if err != nil {
		return errors.Wrap(err, "user update have failed")
	}
	return nil
}

func (m UserMapper) CreateMany(users []models.User) error {
	if len(users) < 1 {
		return nil
	}
	const columnCount = 7
	markStrings := make([]string, 0, len(users)*columnCount)
	valueArgs := make([]interface{}, 0, len(users)*columnCount)

	i := 0
	for _, u := range users {
		markStrings = append(markStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*columnCount+1,
			i*columnCount+2,
			i*columnCount+3,
			i*columnCount+4,
			i*columnCount+5,
			i*columnCount+6,
			i*columnCount+7,
		))
		valueArgs = append(valueArgs,
			u.Username,
			u.FirstName,
			u.LastName,
			u.Email,
			u.Password,
			u.Phone,
			u.UserStatus)
		i++
	}

	stmt := fmt.Sprintf(`INSERT INTO users (username, first_name, last_name, email, password, phone, user_status )
      VALUES %s`, strings.Join(markStrings, ","))
	_, err := m.DB.Exec(stmt, valueArgs...)
	return err
}

func (m UserMapper) DeleteByUsername(username string) error {
	_, err := m.DB.Exec(`DELETE FROM users where username=$1`, username)
	return err
}
