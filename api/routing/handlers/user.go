package handlers

import (
	"gitlab.com/i4s-edu/petstore-kovalyk/services/auth"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/mappers"

	"github.com/go-chi/chi"
	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
	"gitlab.com/i4s-edu/petstore-kovalyk/utils"
)

type User struct {
	UserMapper mappers.UserMapperInterface
}

func (u User) Create(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := &models.User{}
	err = json.Unmarshal(data, &user)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = u.UserMapper.FindByUsername(user.Username)
	logrus.Error(err)
	switch err.(type) {
	case mappers.NotFoundError:
		break
	case error:
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		JSONApiResponse(w, "Username is already in use", http.StatusBadRequest)
		return
	}
	_, err = u.UserMapper.FindByEmail(user.Email)
	switch err.(type) {
	case mappers.NotFoundError:
		break
	case error:
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		JSONApiResponse(w, "Email is already in use", http.StatusBadRequest)
		return
	}
	err = user.Validate()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	hash, err := utils.GetHash(user.Password)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Server error", http.StatusInternalServerError)
		return
	}
	user.Password = hash
	err = u.UserMapper.Create(user)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (u User) CreateWithList(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var users []models.User
	err = json.Unmarshal(data, &users)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, user := range users {
		_, err = u.UserMapper.FindByUsername(user.Username)

		switch err.(type) {
		case mappers.NotFoundError:
			break
		case error:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		default:
			JSONApiResponse(w, "Username is already in use", http.StatusBadRequest)
			return
		}

		_, err = u.UserMapper.FindByEmail(user.Email)
		switch err.(type) {
		case mappers.NotFoundError:
			break
		case error:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		default:
			JSONApiResponse(w, "Email is already in use", http.StatusBadRequest)
			return
		}
		err = user.Validate()
		if err != nil {
			logrus.Error(err)
			JSONApiResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		hash, err2 := utils.GetHash(user.Password)
		if err2 != nil {
			logrus.Error(err)
			JSONApiResponse(w, "Server error", http.StatusInternalServerError)
			return
		}
		user.Password = hash
	}
	err = u.UserMapper.CreateMany(users)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)

}

func (u User) Login(w http.ResponseWriter, r *http.Request) {
	username, err := utils.GetURLParam(r, "username")
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Invalid username/password supplied", http.StatusBadRequest)
	}
	password, err := utils.GetURLParam(r, "password")
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Invalid username/password supplied", http.StatusBadRequest)

	}

	user, err := u.UserMapper.FindByUsername(username)
	if err != nil || !utils.ComparePasswords(password, user.Password) {
		JSONApiResponse(w, "Wrong credentials", http.StatusBadRequest)
		return
	}
	err = auth.GetAuthService().Authenticate(w, r, user)
	if err != nil {
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
	}
}

func (User) Logout(w http.ResponseWriter, r *http.Request) {
	err := auth.GetAuthService().Deauthenticate(r)
	if err != nil {
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
	}
}

func (u User) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := u.UserMapper.FindByUsername(username)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "User not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	output, err := json.Marshal(user)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSONResponse(w, output, http.StatusOK)

}

func (u User) Update(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := u.UserMapper.FindByUsername(username)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "User not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = json.Unmarshal(data, user)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userWithUsername, err := u.UserMapper.FindByUsername(user.Username)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if userWithUsername != nil && user.Username != username {
		JSONApiResponse(w, "Username already in use", http.StatusBadRequest)
		return
	}

	userWithEmail, err := u.UserMapper.FindByEmail(user.Email)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if userWithEmail != nil && userWithEmail.Username != username {
		JSONApiResponse(w, "Email already in use", http.StatusBadRequest)
		return
	}

	err = user.Validate()
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}
	hash, err := utils.GetHash(user.Password)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, "Server error", http.StatusInternalServerError)
		return
	}
	user.Password = hash
	err = u.UserMapper.UpdateByUsername(user, username)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (u User) Delete(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	err := u.UserMapper.DeleteByUsername(username)
	if err != nil {
		logrus.Error(err)
		switch err.(type) {
		case mappers.NotFoundError:
			JSONApiResponse(w, "User not found", http.StatusNotFound)
			return
		default:
			JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}
