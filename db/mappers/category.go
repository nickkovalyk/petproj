package mappers

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
)

type CategoryMapperInterface interface {
	FindByID(id int) (*models.Category, error)
	FindOrCreate(catName string) (*models.Category, error)
	Create(*models.Category) error
	Update(*models.Category) error
	Delete(id int) error
}
type CategoryMapper struct {
	DB *sqlx.DB
	Tx *sqlx.Tx
}

func (m CategoryMapper) FindByID(id int) (*models.Category, error) {
	category := &models.Category{}
	err := m.DB.Get(category, "SELECT * FROM categories where id=$1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NotFoundError("category not found")
		}
		return nil, err
	}
	return category, nil
}

func (m CategoryMapper) FindOrCreate(catName string) (*models.Category, error) {
	category := &models.Category{}
	stmt := `WITH ins AS (
	    	  INSERT INTO categories (name) 
			  VALUES ($1)
			  ON CONFLICT (name) DO NOTHING 
			  RETURNING *
	    	 )
	    	 SELECT * FROM ins
	    	 UNION
	    	 SELECT * FROM categories WHERE name = $2;`
	if m.Tx != nil {
		err := m.Tx.Get(category, stmt, catName, catName)
		if err != nil {
			return nil, errors.Wrap(err, "FindOrCreate category failed")
		}
		return category, nil
	}
	err := m.DB.Get(category, stmt, catName, catName)
	if err != nil {
		return nil, errors.Wrap(err, "FindOrCreate category failed")

	}
	return category, nil

}

func (m CategoryMapper) Create(c *models.Category) error {
	stmt := `INSERT INTO categories ( name ) VALUES ($1) RETURNING id;`
	var categoryID int
	err := m.DB.QueryRowx(stmt, c.Name).Scan(&categoryID)
	if err != nil {
		return errors.Wrap(err, "category create have failed")
	}
	c.ID = categoryID

	return nil
}

func (m CategoryMapper) Update(c *models.Category) error {
	stmt := `UPDATE categories SET name=$1 WHERE id=$2`
	_, err := m.DB.Exec(stmt, c.Name, c.ID)
	if err != nil {
		return errors.Wrap(err, "category update have failed")
	}
	return nil
}

func (m CategoryMapper) Delete(id int) error {
	_, err := m.DB.Exec(`DELETE FROM categories where id=$1`, id)
	return err
}
