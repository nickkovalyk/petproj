package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"gitlab.com/i4s-edu/petstore-kovalyk/db/mappers"
	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
)

type Store struct {
	PetMapper   mappers.PetMapperInterface
	OrderMapper mappers.OrderMapperInterface
}

func (Store) GetInventory(w http.ResponseWriter, r *http.Request) {

	data := struct {
		AdditionalProp1 int
		AdditionalProp2 int
		AdditionalProp3 int
	}{
		1,
		2,
		3,
	}

	output, err := json.Marshal(data)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSONResponse(w, output, http.StatusOK)
}

func (s Store) CreateOrder(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	order := &models.Order{}
	err = json.Unmarshal(b, order)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logrus.Info("order:", order)
	err = order.Validate()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	_, err = s.PetMapper.FindByID(order.PetID)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "Pet with given id have not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err = s.OrderMapper.Create(order)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	output, err := json.Marshal(order)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSONResponse(w, output, http.StatusCreated)

}

func (s Store) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	order, err := s.OrderMapper.FindByID(id)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "Order not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	output, err := json.Marshal(order)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSONResponse(w, output, http.StatusOK)

}

func (s Store) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		JSONApiResponse(w, "Invalid ID supplied", http.StatusBadRequest)
		return
	}

	err = s.OrderMapper.Delete(id)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "Order not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}
