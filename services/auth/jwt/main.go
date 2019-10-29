package jwt

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"

	jwt "github.com/dgrijalva/jwt-go"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type AuthService struct {
	claims *Claims
	jwtKey []byte
	mx     *sync.Mutex
	users  map[string]*models.User
}

const tokenKey = "token"
const expireTime = 10 * time.Minute

func NewJwtAuthService() AuthService {
	authService := AuthService{
		mx:     &sync.Mutex{},
		jwtKey: []byte("secret"),
		claims: &Claims{},
		users:  map[string]*models.User{},
	}
	authService.cleaner()
	return authService
}
func (jwa AuthService) Authenticate(w http.ResponseWriter, r *http.Request, user *models.User) error {

	if jwa.IsAuthenticated(r) {
		err := jwa.Deauthenticate(r)
		if err != nil {
			return err
		}
	}
	expiresAt := time.Now().Add(expireTime)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwa.jwtKey)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:    tokenKey,
		Value:   tokenString,
		Expires: expiresAt,
	})
	jwa.mx.Lock()
	jwa.users[tokenString] = user
	jwa.mx.Unlock()
	w.Header().Set("X-Expires-After", string(expiresAt.Unix()))

	return nil
}

func (jwa AuthService) Deauthenticate(r *http.Request) error {
	c, err := r.Cookie(tokenKey)
	if err != nil {
		return errors.New("token is not present  ")
	}
	jwa.mx.Lock()
	delete(jwa.users, c.Value)
	jwa.mx.Unlock()
	return nil
}

func (jwa AuthService) RefreshAuth(w http.ResponseWriter, r *http.Request) error {
	user := jwa.GetUser(r)
	if user == nil {
		return errors.New("no user present in system")
	}
	if err := jwa.Deauthenticate(r); err != nil {
		return err
	}
	err := jwa.Authenticate(w, r, user)
	if err != nil {
		return err
	}
	return nil
}
func (jwa AuthService) GetUser(r *http.Request) *models.User {
	if jwa.CheckAuth(r) != http.StatusOK {
		return nil
	}
	c, err := r.Cookie(tokenKey)
	if err != nil {
		return nil
	}
	tokenString := c.Value
	jwa.mx.Lock()
	user := jwa.users[tokenString]
	jwa.mx.Unlock()
	return user
}

func (jwa AuthService) IsAuthenticated(r *http.Request) bool {
	if jwa.CheckAuth(r) != http.StatusOK {
		return false
	}
	c, err := r.Cookie(tokenKey)
	if err != nil {
		return false
	}
	jwa.mx.Lock()
	_, ok := jwa.users[c.Value]
	jwa.mx.Unlock()

	return ok
}

func (jwa AuthService) CheckAuth(r *http.Request) (httpStatusCode int) {
	c, err := r.Cookie(tokenKey)
	if err != nil {
		if err == http.ErrNoCookie {
			return http.StatusUnauthorized
		}
		return http.StatusBadRequest
	}
	tokenString := c.Value
	token, err := jwt.ParseWithClaims(tokenString, jwa.claims, func(token *jwt.Token) (interface{}, error) {
		return jwa.jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return http.StatusUnauthorized
		}
		return http.StatusBadRequest
	}
	if !token.Valid {
		return http.StatusUnauthorized
	}
	return http.StatusOK
}

func (jwa AuthService) cleaner() {
	go func() {
		var err error
		for {
			jwa.mx.Lock()
			for tokenString := range jwa.users {
				_, err = jwt.ParseWithClaims(tokenString, jwa.claims, func(token *jwt.Token) (interface{}, error) {
					return jwa.jwtKey, nil
				})
				if err != nil {
					delete(jwa.users, tokenString)
				}
			}
			jwa.mx.Unlock()
			time.Sleep(expireTime)
		}
	}()
}
