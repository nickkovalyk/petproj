package mappers

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
)

type PetMapperInterface interface {
	FindByID(id int) (*models.Pet, error)
	FindByStatus(status string) ([]*models.Pet, error)
	FindByTags(tags []string) ([]*models.Pet, error)
	Create(*models.Pet) error
	Update(*models.Pet) error
	Delete(id int) error
}

type PetMapper struct {
	DB *sqlx.DB
}

func (m PetMapper) FindByID(id int) (*models.Pet, error) {
	stmt := `
		SELECT p.id, p.name, p.status, p.photo_urls, c.id, c.id, c.name, t.id, t.name  FROM pets p
		    LEFT JOIN categories c ON p.category_id = c.id
		    LEFT JOIN pet_tag pt ON p.id = pt.pet_id
		    INNER JOIN tags t ON pt.tag_id = t.id
		WHERE p.id = $1`

	p := &models.Pet{}
	rows, err := m.DB.Queryx(stmt, id)
	if err != nil {
		return nil, errors.Wrap(err, "find pet by id error ")
	}
	tag := models.Tag{}
	for rows.Next() {
		err = rows.Scan(
			&p.ID,
			&p.Name,
			&p.Status,
			pq.Array(&p.PhotoURLs),
			&p.CategoryID,
			&p.Category.ID,
			&p.Category.Name,
			&tag.ID,
			&tag.Name,
		)
		p.Tags = append(p.Tags, tag)
		if err != nil {
			return nil, errors.Wrap(err, "scan pet error ")
		}
	}
	logrus.Info(p)

	if p.ID == 0 {
		return p, NotFoundError(fmt.Sprintf("Pet record have not found by id: %d", id))
	}

	return p, nil
}

func (m PetMapper) Create(p *models.Pet) error {
	stmt := `INSERT INTO pets ( name, status, photo_urls, category_id) 
			 VALUES (:name, :status, :photo_urls, :category_id) RETURNING id;`
	txn, err := m.DB.Beginx()
	defer func() {
		if err = txn.Rollback(); err != sql.ErrTxDone {
			logrus.Error("rollback error", err)
		}
	}()
	if err != nil {
		return errors.Wrap(err, "transaction open err")
	}
	logrus.Info(p)
	category, err := CategoryMapper{Tx: txn}.FindOrCreate(p.Category.Name)
	if err != nil {
		return errors.Wrap(err, "can not find category")
	}
	tags, err := TagMapper{Tx: txn}.FindOrCreateMany(p.Tags)
	if err != nil {
		return errors.Wrap(err, "problem with tags")
	}

	//main query
	var petID int
	params := map[string]interface{}{
		"name":        p.Name,
		"status":      p.Status,
		"photo_urls":  pq.Array(p.PhotoURLs),
		"category_id": category.ID,
	}
	rows, err := txn.NamedQuery(stmt, params)
	if err != nil {
		return errors.Wrap(err, "insert pet error")
	}
	for rows.Next() {
		err = rows.Scan(&petID)
		if err != nil {
			return errors.Wrap(err, "scan pet id error")
		}
	}
	logrus.Infof("petID:%v", petID)

	err = m.DissociateAllTags(txn, petID)
	if err != nil {
		return errors.Wrap(err, "tag dissociate fail")
	}

	err = m.AssociateTags(txn, petID, tags)
	if err != nil {
		return errors.Wrap(err, "tag associate fail")
	}

	err = txn.Commit()
	if err != nil {
		return errors.Wrap(err, "Transaction commit fail")
	}
	return nil
}

func (m PetMapper) Update(p *models.Pet) error {
	stmt := `UPDATE pets SET name=:name, status=:status, photo_urls=:photo_urls, category_id=:category_id
             WHERE id=:id`
	txn, err := m.DB.Beginx()
	defer func() {
		if err = txn.Rollback(); err != sql.ErrTxDone {
			logrus.Error("rollback error", err)
		}
	}()
	if err != nil {
		return errors.Wrap(err, "transaction open error")
	}

	category, err := CategoryMapper{Tx: txn}.FindOrCreate(p.Category.Name)
	if err != nil {
		return errors.Wrap(err, "can not find category")
	}

	tags, err := TagMapper{Tx: txn}.FindOrCreateMany(p.Tags)
	if err != nil {
		return errors.Wrap(err, "problem with tags")
	}

	// main query
	params := map[string]interface{}{
		"name":        p.Name,
		"status":      p.Status,
		"photo_urls":  pq.Array(p.PhotoURLs),
		"category_id": category.ID,
		"id":          p.ID,
	}
	_, err = txn.NamedExec(stmt, params)
	if err != nil {
		return errors.Wrap(err, "db exec fail")
	}
	err = m.DissociateAllTags(txn, p.ID)
	if err != nil {
		return errors.Wrap(err, "tag dissociate fail")
	}

	err = m.AssociateTags(txn, p.ID, tags)
	if err != nil {
		return errors.Wrap(err, "tag associate fail")
	}

	err = txn.Commit()
	if err != nil {
		return errors.Wrap(err, "Transaction commit fail")
	}

	return nil
}

func (m PetMapper) Delete(id int) error {
	stmt := `DELETE from pets WHERE id=$1`
	_, err := m.DB.Exec(stmt, id)
	return err
}

func (m PetMapper) FindByStatus(status string) ([]*models.Pet, error) {
	stmt := `
		 SELECT p.id, p.name, p.status, p.photo_urls, c.id, c.id, c.name, t.id, t.name  
		 	FROM pets p
		 	LEFT JOIN categories c ON p.category_id = c.id
		 	LEFT JOIN pet_tag pt ON p.id = pt.pet_id
		 	INNER JOIN tags t ON pt.tag_id = t.id
		 WHERE status = $1`

	rows, err := m.DB.Queryx(stmt, status)
	if err != nil {
		return nil, errors.Wrap(err, "find by tags error ")
	}
	return m.MapsRows(rows)
}
func (m PetMapper) FindByTags(tags []string) ([]*models.Pet, error) {
	stmt := `
		 SELECT p.id, p.name, p.status, p.photo_urls, c.id, c.id, c.name, t.id, t.name  
		 	FROM pets p
		 	LEFT JOIN categories c ON p.category_id = c.id
		 	LEFT JOIN pet_tag pt ON p.id = pt.pet_id
		 	INNER JOIN tags t ON pt.tag_id = t.id
		 WHERE $1 <@  array(SELECT t.name 
		 						FROM pets  
		     					LEFT JOIN pet_tag pt ON p.id = pt.pet_id
		     					INNER JOIN tags t ON pt.tag_id = t.id)`

	rows, err := m.DB.Queryx(stmt, pq.Array(tags))
	if err != nil {
		return nil, errors.Wrap(err, "find by tags error ")
	}
	return m.MapsRows(rows)

}

func (m PetMapper) AssociateTags(txn *sqlx.Tx, petID int, tags []models.Tag) error {
	if len(tags) < 1 {
		return nil
	}
	const columnCount = 2
	markStrings := make([]string, 0, len(tags)*columnCount)
	valueArgs := make([]interface{}, 0, len(tags)*columnCount)
	i := 0
	for _, tag := range tags {
		markStrings = append(markStrings, fmt.Sprintf("($%d, $%d)", i*columnCount+1, i*columnCount+2))
		valueArgs = append(valueArgs, petID, tag.ID)
		i++
	}
	stmt := fmt.Sprintf("INSERT INTO pet_tag (pet_id, tag_id) VALUES %s", strings.Join(markStrings, ","))
	_, err := txn.Exec(stmt, valueArgs...)
	return err

}

func (m PetMapper) DissociateAllTags(txn *sqlx.Tx, petID int) error {
	_, err := txn.Exec(`DELETE FROM pet_tag where pet_id=$1`, petID)
	return err
}

func (PetMapper) MapsRows(rows *sqlx.Rows) ([]*models.Pet, error) {
	p := &models.Pet{}
	tag := models.Tag{}
	petsMap := map[int]*models.Pet{}
	for rows.Next() {
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Status,
			pq.Array(&p.PhotoURLs),
			&p.CategoryID,
			&p.Category.ID,
			&p.Category.Name,
			&tag.ID,
			&tag.Name,
		)
		if err != nil {
			return nil, errors.Wrap(err, "scan pet error ")
		}

		_, present := petsMap[p.ID]
		if !present {
			pet := *p
			petsMap[p.ID] = &pet
		}
		petsMap[p.ID].Tags = append(petsMap[p.ID].Tags, tag)

	}
	var pets = make([]*models.Pet, 0, len(petsMap))
	for _, pet := range petsMap {
		pets = append(pets, pet)
	}
	return pets, nil
}
