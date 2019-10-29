package auth

import (
	"net/http"

	jwt2 "gitlab.com/i4s-edu/petstore-kovalyk/services/auth/jwt"

	"github.com/sirupsen/logrus"
	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
)

type Config struct {
	Type string
}

type ServiceInterface interface {
	Authenticate(w http.ResponseWriter, r *http.Request, user *models.User) error
	Deauthenticate(r *http.Request) error
	RefreshAuth(w http.ResponseWriter, r *http.Request) error
	GetUser(r *http.Request) *models.User
	IsAuthenticated(r *http.Request) bool
	CheckAuth(r *http.Request) (httpStatusCode int)
}

var service ServiceInterface

func Init(config Config) {
	switch config.Type {
	case "jwt":
		service = jwt2.NewJwtAuthService()
		logrus.Info("JWT auth service initialized")
	default:
		logrus.Fatalf("unsupported auth type")
	}
}

func GetAuthService() ServiceInterface {
	if service == nil {
		logrus.Fatalf("authService has not initialized")
	}
	return service
}
