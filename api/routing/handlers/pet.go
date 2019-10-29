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
	"gitlab.com/i4s-edu/petstore-kovalyk/utils"
)

const imagePath = "static/uploads/images"

type Pet struct {
	PetMapper mappers.PetMapperInterface
}

func (p Pet) Create(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logrus.Errorf("Body reading error: %v", err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pet := &models.Pet{}
	err = json.Unmarshal(b, pet)
	if err != nil {
		logrus.Errorf("Pet model unmarshal has failed: %v", err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = pet.Validate()
	if err != nil {
		logrus.Error("pet validation error", err)
		JSONApiResponse(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	err = p.PetMapper.Create(pet)
	if err != nil {
		logrus.Error("Pet model save has failed", err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)

}

func (p Pet) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		logrus.Error("Failed convert id param to int", err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if id < 1 {
		JSONApiResponse(w, "Invalid id supplied", http.StatusBadRequest)
		return
	}
	pet, err := p.PetMapper.FindByID(id)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "Pet not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	output, err := json.Marshal(pet)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSONResponse(w, output, http.StatusOK)

}

func (p Pet) Update(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pet := &models.Pet{}
	err = json.Unmarshal(data, pet)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pet.ID < 1 {
		JSONApiResponse(w, "Invalid ID supplied", http.StatusBadRequest)
		return
	}

	_, err = p.PetMapper.FindByID(pet.ID)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "Pet not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = pet.Validate()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	err = p.PetMapper.Update(pet)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (p Pet) FindByStatus(w http.ResponseWriter, r *http.Request) {

	status, err := utils.GetURLParam(r, "status")
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Invalid status value", http.StatusBadRequest)
		return
	}
	err = models.Pet{}.CheckStatus(status)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	pets, err := p.PetMapper.FindByStatus(status)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	output, err := json.Marshal(pets)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSONResponse(w, output, http.StatusOK)

}

func (p Pet) FindByTags(w http.ResponseWriter, r *http.Request) {

	tags, err := utils.GetURLParams(r, "tags")
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Invalid tag value", http.StatusBadRequest)
		return
	}
	err = models.Pet{}.CheckTags(tags)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Invalid tag value", http.StatusBadRequest)
		return
	}

	pets, err := p.PetMapper.FindByTags(tags)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	output, err := json.Marshal(pets)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSONResponse(w, output, http.StatusOK)

}

func (p Pet) UpdateByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		JSONApiResponse(w, "Invalid ID supplied", http.StatusBadRequest)
		return
	}

	pet, err := p.PetMapper.FindByID(id)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "Pet not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if name := utils.GetFormParam(r, "name"); len(name) > 0 {
		pet.Name = name
	}
	if status := utils.GetFormParam(r, "status"); len(status) > 0 {
		pet.Status = status
	}

	err = pet.Validate()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Invalid input", http.StatusMethodNotAllowed)
		return
	}

	err = p.PetMapper.Update(pet)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (p Pet) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		JSONApiResponse(w, "Invalid ID supplied", http.StatusBadRequest)
		return
	}

	err = p.PetMapper.Delete(id)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "Pet not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func (p Pet) UploadImage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		JSONApiResponse(w, "Invalid ID supplied", http.StatusBadRequest)
		return
	}

	pet, err := p.PetMapper.FindByID(id)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "Pet not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	allowedMIMETypes := []string{"image/jpeg", "image/png"}

	filename := utils.RandomString(16)
	imgURL, err := utils.SaveFile(r, "file", imagePath, filename, allowedMIMETypes)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Server error", http.StatusInternalServerError)
		return
	}
	pet.PhotoURLs = append(pet.PhotoURLs, imgURL)

	err = p.PetMapper.Update(pet)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	JSONApiResponse(w, "Success upload", http.StatusOK)
}
