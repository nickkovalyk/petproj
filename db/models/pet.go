package models

import (
	"github.com/sirupsen/logrus"

	"gitlab.com/i4s-edu/petstore-kovalyk/utils"
)

var allowedPetStatuses = []string{"available", "pending", "sold"}

type Pet struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	PhotoURLs  []string `json:"photoUrls" db:"photo_urls"`
	Tags       []Tag    `json:"tags"`
	Category   Category `json:"category"`
	CategoryID int      `json:"-" db:"category_id"`
}

func (p *Pet) Validate() error {
	err := p.CheckStatus(p.Status)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (Pet) CheckStatus(status string) error {
	if utils.ContainsString(status, allowedPetStatuses) {
		return nil
	}
	return ValidationError("not allowed status for pet model")
}
func (Pet) CheckTags(tags []string) error {
	return nil
}
