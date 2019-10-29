package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"gitlab.com/i4s-edu/petstore-kovalyk/utils"
)

var allowedOrderStatuses = []string{"placed", "approved", "delivered"}

//"2019-08-26T11:55:56.457Z"
const ShipDateFormat = "2006-01-02T15:04:05.999Z0700"

type Order struct {
	ID       int    `json:"id"`
	PetID    int    `json:"petId" db:"pet_id"`
	Quantity int    `json:"quantity"`
	ShipDate string `json:"shipDate" db:"ship_date"`
	Complete bool   `json:"complete"`
	Status   string `json:"status"`
}

func (o *Order) MarshalJSON() (output []byte, err error) {
	type Alias Order
	i, err := strconv.ParseInt(o.ShipDate, 10, 64)
	if err != nil {
		return
	}
	unixTimeUTC := time.Unix(i, 0)
	unitTimeInRFC3339 := unixTimeUTC.Format(ShipDateFormat)
	o.ShipDate = unitTimeInRFC3339
	return json.Marshal((*Alias)(o))
}

func (o *Order) UnmarshalJSON(bytes []byte) error {
	type Alias Order
	alias := &Alias{}
	err := json.Unmarshal(bytes, alias)
	if err != nil {
		return err
	}
	t, err := time.Parse(ShipDateFormat, alias.ShipDate)
	if err != nil {
		return err
	}
	alias.ShipDate = fmt.Sprintf("%d", t.Unix())
	*o = *((*Order)(alias))

	return nil
}

func (o *Order) Validate() error {
	err := o.checkStatus()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if o.Quantity < 1 {
		return errors.New("invalid quantity")
	}
	return nil
}

func (o *Order) checkStatus() error {
	if utils.ContainsString(o.Status, allowedOrderStatuses) {
		return nil
	}
	return ValidationError("not allowed status for order model")
}
