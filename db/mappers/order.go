package mappers

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"time"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
)

type OrderMapperInterface interface {
	FindByID(id int) (*models.Order, error)
	GetOldest(int64) ([]*models.Order, error)
	Create(o *models.Order) error
	Update(o *models.Order) error
	Delete(id int) error
}
type OrderMapper struct {
	DB *sqlx.DB
}

func (m OrderMapper) FindByID(id int) (*models.Order, error) {
	order := &models.Order{}
	err := m.DB.Get(order, "SELECT * FROM orders where id=$1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NotFoundError("order not found")
		}
		return nil, err
	}
	return order, nil
}

func (m OrderMapper) GetOldest(timestamp int64) (orders []*models.Order, err error) {
	stmt := "SELECT * FROM orders where ship_date > $1"
	err = m.DB.Select(&orders, stmt, timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NotFoundError("orders not found")
		}
	}
	return
}

func (m OrderMapper) Create(o *models.Order) error {
	stmt := `INSERT INTO orders ( pet_id, quantity, ship_date, complete, status) 
             VALUES (:pet_id, :quantity, :ship_date, :complete, :status)
             RETURNING id;`
	timestamp := time.Now().Unix()
	var orderID int
	params := map[string]interface{}{
		"pet_id":    o.PetID,
		"quantity":  o.Quantity,
		"ship_date": timestamp,
		"complete":  o.Complete,
		"status":    o.Status,
	}
	rows, err := m.DB.NamedQuery(stmt, params)
	if err != nil {
		return errors.Wrap(err, "insert order error")
	}
	for rows.Next() {
		err := rows.Scan(&orderID)
		if err != nil {
			return errors.Wrap(err, "scan order id error")
		}
	}
	o.ID = orderID
	return nil
}

func (m OrderMapper) Update(o *models.Order) error {
	stmt := `UPDATE orders SET pet_id=:pet_id, quantity=:quantity, ship_date=:ship_date, 
                  complete=:complete, status=:status WHERE id=:id`
	timestamp := time.Now().Unix()
	params := map[string]interface{}{
		"pet_id":    o.PetID,
		"quantity":  o.Quantity,
		"ship_date": timestamp,
		"complete":  o.Complete,
		"status":    o.Status,
		"id":        o.ID,
	}
	_, err := m.DB.NamedExec(stmt, params)
	if err != nil {
		return errors.Wrap(err, "order update failed")
	}

	return nil
}

func (m OrderMapper) Delete(id int) error {
	_, err := m.DB.Exec(`DELETE FROM orders where id=$1`, id)
	return err
}
