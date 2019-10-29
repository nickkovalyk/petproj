package utils

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/sirupsen/logrus"
)

var extensions = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func ContainsString(needle string, data []string) bool {
	for _, v := range data {
		if v == needle {
			return true
		}
	}
	return false
}

func GetURLParam(r *http.Request, parameterName string) (string, error) {
	keys, ok := r.URL.Query()[parameterName]

	if !ok || len(keys[0]) < 1 {
		return "", errors.New("parameter access error")
	}
	return keys[0], nil
}

func GetURLParams(r *http.Request, paramName string) ([]string, error) {
	keys, ok := r.URL.Query()[paramName]

	if !ok || len(keys[0]) < 1 {
		return []string{}, errors.New("parameter access error")
	}
	return keys, nil
}

func GetFormParam(r *http.Request, paramName string) string {
	err := r.ParseForm()
	if err != nil {
		logrus.Error(err)
		return ""
	}
	return r.Form.Get(paramName)
}

func SaveFile(r *http.Request, fileKey, path, filename string, allowedMimeTypes []string) (string, error) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return "", err
	}
	file, handler, err := r.FormFile(fileKey)
	if err != nil {
		return "", err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			logrus.Error(err)
		}
	}()

	mimeType := handler.Header.Get("Content-Type")
	if !ContainsString(mimeType, allowedMimeTypes) {
		return "", errors.New("not allowed mimetype")

	}
	extension, ok := extensions[mimeType]
	if !ok {
		return "", errors.New("unknown mimetype")
	}

	fullPath := fmt.Sprintf("%v/%v.%v", path, filename, extension)
	f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			logrus.Error(err)
		}
	}()

	_, err = io.Copy(f, file)
	if err != nil {
		return "", err
	}
	return fullPath, err
}

func GetHash(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
func ComparePasswords(pwd, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits

	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	return string(buf)
}
