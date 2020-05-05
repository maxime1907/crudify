package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/maxime1907/crudify/config"
	"github.com/maxime1907/crudify/logger"
	"github.com/maxime1907/crudify/handler"
)

func Validate(username string, password string, routerinfo config.RouterInfo) bool {
    if username == routerinfo.Username && password == routerinfo.Password {
        return true
    }
    return false
}

func BasicAuth(inner http.Handler, routerinfo config.RouterInfo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger.Log(r).Debug().Msg("Verifying authentification")

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

        if len(auth) != 2 || auth[0] != "Basic" {
			err := errors.New("Authorization failed : not a basic auth")
			logger.Log(r).Warn().Msg(err.Error())
			err = handler.SendAnswer(w, r, nil, err)
			if err != nil {
				logger.Log(r).Warn().Msg(err.Error())
			}
            return
        }

        payload, _ := base64.StdEncoding.DecodeString(auth[1])
        pair := strings.SplitN(string(payload), ":", 2)

        if len(pair) != 2 || !Validate(pair[0], pair[1], routerinfo) {
			err := errors.New("Authorization failed : username and password does not match")
			logger.Log(r).Warn().Msg(err.Error())
			err = handler.SendAnswer(w, r, nil, err)
			if err != nil {
				logger.Log(r).Warn().Msg(err.Error())
			}
            return
		}

		inner.ServeHTTP(w, r)
	})
}