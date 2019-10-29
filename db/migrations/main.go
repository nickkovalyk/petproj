package migrations

import (
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func Run(db *sqlx.DB) error {
	logrus.Info("Start migrations")

	err := createUsersTable(db)
	if err != nil {
		return err
	}
	err = createCategoriesTable(db)
	if err != nil {
		return err
	}
	err = createPetsTable(db)
	if err != nil {
		return err
	}
	err = createOrdersTable(db)
	if err != nil {
		return err
	}
	err = createTagsTable(db)
	if err != nil {
		return err
	}
	err = createPetTagTable(db)
	if err != nil {
		return err
	}
	err = createInvoicesTable(db)
	if err != nil {
		return err
	}
	logrus.Info("Successfully migrated")

	return nil
}

func createUsersTable(db *sqlx.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS  users (
			    id SERIAL PRIMARY KEY,
			    username VARCHAR(255) UNIQUE,
			    first_name VARCHAR(255),
			    last_name VARCHAR(255),
			    email TEXT UNIQUE NOT NULL,
			    password VARCHAR(255),
			    phone VARCHAR(15),
			    user_status INT
			 );`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func createCategoriesTable(db *sqlx.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS categories (
			    id SERIAL PRIMARY KEY,
			    name VARCHAR(255) UNIQUE
			 );`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func createPetsTable(db *sqlx.DB) error {

	stmt := `CREATE TABLE IF NOT EXISTS pets ( 
			    id SERIAL PRIMARY KEY,
			    category_id INT references categories(id) ON DELETE CASCADE,
			    name VARCHAR(255),
			    photo_urls text[],
			    status VARCHAR(255)
			 );`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}
func createOrdersTable(db *sqlx.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS orders (
			    id SERIAL PRIMARY KEY,
			    quantity INT,
			    ship_date BIGINT,
			    complete VARCHAR(255),
			    status VARCHAR(255),
			    pet_id INT references pets(id) ON DELETE CASCADE
			 );`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func createTagsTable(db *sqlx.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS tags (
			    id SERIAL PRIMARY KEY,
			    name VARCHAR(255) UNIQUE 
			 );`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func createPetTagTable(db *sqlx.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS pet_tag (
			    pet_id int references pets(id) ON DELETE CASCADE,
			    tag_id int references tags(id) ON DELETE CASCADE,
			    UNIQUE(pet_id, tag_id) 
			 );`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func createInvoicesTable(db *sqlx.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS invoices (
			   id SERIAL PRIMARY KEY,
			   body TEXT NOT NULL ,
			   created_date BIGINT NOT NULL 
             );`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}
