package mappers

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
)

type TagMapperInterface interface {
	FindByID(id int) (*models.Tag, error)
	Create(*models.Tag) error
	Update(*models.Tag) error
	Delete(id int) error
}
type TagMapper struct {
	DB *sqlx.DB
	Tx *sqlx.Tx
}

func (m TagMapper) FindByID(id int) (*models.Tag, error) {
	tag := &models.Tag{}
	err := m.DB.Get(tag, "SELECT * FROM tags where id=$1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NotFoundError("tag not found")
		}
		return nil, err
	}
	return tag, nil
}

func (m TagMapper) Create(t *models.Tag) error {
	stmt := `INSERT INTO tags ( name ) VALUES ($1) RETURNING id;`
	var tagID int
	err := m.DB.QueryRowx(stmt, t.Name).Scan(&tagID)
	if err != nil {
		return errors.Wrap(err, "tag create have failed")
	}
	t.ID = tagID

	return nil
}

func (m TagMapper) Update(t *models.Tag) error {
	stmt := `UPDATE tags SET name=$1 WHERE id=$2`
	_, err := m.DB.Exec(stmt, t.Name, t.ID)
	if err != nil {
		return errors.Wrap(err, "tag update have failed")
	}
	return nil
}

func (m TagMapper) FindOrCreateMany(tags []models.Tag) (dbTags []models.Tag, err error) {
	if len(tags) < 1 {
		return
	}
	var markStrings []string
	var inStmtStrings []string
	var valueArgs []interface{}
	i := 1
	for _, tag := range tags {
		markStrings = append(markStrings, fmt.Sprintf("($%d)", i))
		inStmtStrings = append(inStmtStrings, fmt.Sprintf("$%d", i+len(tags)))
		valueArgs = append(valueArgs, tag.Name)
		i++
	}
	stmt := fmt.Sprintf(
		`WITH ins AS (
	    	  INSERT INTO tags (name) 
			  VALUES %v
			  ON CONFLICT (name) DO NOTHING 
			  RETURNING *
	    	 )
	    	 SELECT * FROM ins
	    	 UNION
	    	 SELECT * FROM tags WHERE name IN(%v);`,
		strings.Join(markStrings, ","), strings.Join(inStmtStrings, ","))

	if m.Tx != nil {
		err = m.Tx.Select(&dbTags, stmt, append(valueArgs, valueArgs...)...)
		return
	}
	err = m.DB.Select(&dbTags, stmt, append(valueArgs, valueArgs...)...)
	return
}

func (m TagMapper) Delete(id int) error {
	_, err := m.DB.Exec(`DELETE FROM tags where id=$1`, id)
	return err
}
