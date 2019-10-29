package mappers

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
)

type InvoiceMapperInterface interface {
	GetLast() (*models.Invoice, error)
	FindByID(id int) (*models.Invoice, error)
	Create(*models.Invoice) error
	Update(*models.Invoice) error
	Delete(id int) error
}
type InvoiceMapper struct {
	DB *sqlx.DB
	Tx *sqlx.Tx
}

func (m InvoiceMapper) GetLast() (*models.Invoice, error) {
	invoice := &models.Invoice{}
	err := m.DB.Get(invoice, "SELECT * FROM invoices ORDER BY id LIMIT 1")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NotFoundError("invoice not found")
		}
		return nil, err
	}
	return invoice, nil
}

func (m InvoiceMapper) FindByID(id int) (*models.Invoice, error) {
	invoice := &models.Invoice{}
	err := m.DB.Get(invoice, "SELECT * FROM invoices where id=$1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NotFoundError("invoice not found")
		}
		return nil, err
	}
	return invoice, nil
}

func (m InvoiceMapper) Create(i *models.Invoice) error {
	stmt := `INSERT INTO invoices ( body, created_date ) VALUES ($1, $2) RETURNING id;`
	var invoiceID int
	err := m.DB.QueryRowx(stmt, i.Body, i.CreatedDate).Scan(&invoiceID)
	if err != nil {
		return errors.Wrap(err, "invoice create have failed")
	}
	i.ID = invoiceID

	return nil
}

func (m InvoiceMapper) Update(i *models.Invoice) error {
	stmt := `UPDATE invoices SET body=$1 WHERE created_date=$2`
	_, err := m.DB.Exec(stmt, i.Body, i.CreatedDate)
	if err != nil {
		return errors.Wrap(err, "invoice update have failed")
	}
	return nil
}

func (m InvoiceMapper) Delete(id int) error {
	_, err := m.DB.Exec(`DELETE FROM invoices where id=$1`, id)
	return err
}
