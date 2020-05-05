package handler

import (
	"crypto/tls"
	"net/smtp"
	"net/mail"
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"
	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"github.com/maxime1907/crudify/dbhelper"
	"github.com/maxime1907/crudify/logger"
	"github.com/maxime1907/crudify/config"
)

type Response struct {
	Uuid    string      `json:"uuid"`
	Time    time.Time   `json:"time"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func VerifyHash(plainContent string, hashedContent string) error {
	myPlainContent := []byte(plainContent)
    byteHash := []byte(hashedContent)
    err := bcrypt.CompareHashAndPassword(byteHash, myPlainContent)
    return err
}

func HashAndSalt(content string) (string, error) {
	myContent := []byte(content)
    hash, err := bcrypt.GenerateFromPassword(myContent, bcrypt.MinCost)
    return string(hash), err
}

//Get a http status code by looking at an error
func GetStatusCode(r *http.Request, er error) int {
	logger.Log(r).Debug().Msg("Getting status code by error")
	if er != nil {
		var erStr = er.Error()
		if strings.Contains(erStr, "sql: no rows in result set") {
			return http.StatusNotFound
		}
		if strings.Contains(erStr, "violates") {
			return http.StatusConflict
		}
		if strings.Contains(erStr, "Authorization failed") {
			return http.StatusUnauthorized
		}
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

//Sends an HTTP answer with any data you need to pass
func SendAnswer(w http.ResponseWriter, r *http.Request, data interface{}, er error) error {
	var statuscode int = GetStatusCode(r, er)
	var myuuid string
	var msg string

	logger.Log(r).Debug().Msg("Sending HTTP answer with status code " + strconv.Itoa(statuscode))

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statuscode)
	if er != nil {
		msg = er.Error()
	}
	if r != nil {
		if myuuidVal := r.Context().Value("uuid"); myuuidVal != nil {
			myuuid = myuuidVal.(string)
		}
	}
	res := Response{Uuid: myuuid, Time: time.Now(), Message: msg, Data: data}
	err := EncodeJSON(w, r, res)
	return err
}

// Get route parameters
func Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// Get table name from route url
func GetTableName(r *http.Request) string {
	logger.Log(r).Debug().Msg("Getting table name")
	name := r.URL.Path[1:]
	pos := strings.Index(name, "/")
	if pos > -1 {
		name = name[0:pos]
	}
//	name = "\"" + name + "\""
	return name
}

// Get values passed in url as query arguments
func FormToMap(r *http.Request) map[string]string {
	logger.Log(r).Debug().Msg("Mapping form parameters")

	var mymap = map[string]string{}

	r.FormValue("")
	for key, value := range r.Form {
		mymap[key] = value[0]
	}
	return mymap
}

func CheckProperties(data map[string]interface{}, properties []string) error {
	for i := 0; i < len(properties); i++ {
		if _, ok := data[properties[i]]; !ok {
			return errors.New("Missing property " + properties[i]);
		}
	}
	return nil
}

func SendEmail(smtpConfig config.SMTPInfo, email string, title string, body string) error {
	ownerMail := smtpConfig.OwnerEmail
	password := smtpConfig.Password

	// Connect to the SMTP Server
	servername := smtpConfig.Host
	serverport := fmt.Sprintf("%v", smtpConfig.Port)

	from := mail.Address{"", ownerMail}
    to   := mail.Address{"", email }

    // Setup headers
    headers := make(map[string]string)
    headers["From"] = from.String()
    headers["To"] = to.String()
    headers["Subject"] = title

    // Setup message
    message := ""
    for k,v := range headers {
        message += fmt.Sprintf("%s: %s\r\n", k, v)
    }
    message += "\r\n" + body

	auth := smtp.PlainAuth(
		"",
		ownerMail,
		password,
		servername,
	)

    // TLS config
    tlsconfig := &tls.Config {
        InsecureSkipVerify: false,
        ServerName: servername,
    }

    c, err := smtp.Dial(servername + ":" + serverport)
    if err != nil {
        return err
    }

    c.StartTLS(tlsconfig)

    // Auth
    if err = c.Auth(auth); err != nil {
        return err
    }

    // To && From
    if err = c.Mail(from.Address); err != nil {
        return err
    }

    if err = c.Rcpt(to.Address); err != nil {
        return err
    }

    // Data
    w, err := c.Data()
    if err != nil {
        return err
    }

    _, err = w.Write([]byte(message))
    if err != nil {
        return err
    }

    err = w.Close()
    if err != nil {
        return err
    }

	c.Quit()

	return nil
}

func EncodeJSON(w http.ResponseWriter, r *http.Request, res interface{}) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	logger.Log(r).Debug().Msg("Encoding response as a JSON")

	return json.NewEncoder(w).Encode(res)
}

func DecodeJSON(r *http.Request) (*[]map[string]interface{}, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var data []map[string]interface{}
	var check map[string]interface{}

	logger.Log(r).Debug().Msg("Parsing body as a JSON")

	if r != nil && r.Body != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		responseString := buf.String()
		err := json.NewDecoder(buf).Decode(&check)
		if err != nil && err.Error() != "EOF" && len(check) <= 0 {
			err = json.NewDecoder(bytes.NewBufferString(responseString)).Decode(&data)
		} else {
			data = append(data, check)
		}
		if err != nil && err.Error() != "EOF" {
			return nil, err
		}
	} else {
		return nil, errors.New("HTTP body does not exist")
	}
	return &data, nil
}

// Generic get
func Get(w http.ResponseWriter, r *http.Request) {
	args := FormToMap(r)
	tablename := GetTableName(r)
	result, err := dbhelper.Select(r, tablename, args)
	if err != nil {
		logger.Log(r).Warn().Msg(err.Error())
	}
	err = SendAnswer(w, r, result, err)
	if err != nil {
		logger.Log(r).Warn().Msg(err.Error())
	}
}

// Generic post
func Post(w http.ResponseWriter, r *http.Request) {
	var result *[]map[string]interface{}

	args := FormToMap(r)
	tablename := GetTableName(r)
	data, err := DecodeJSON(r)
	if err == nil {
		result, err = dbhelper.Insert(r, tablename, args, *data)
	}
	if err != nil {
		logger.Log(r).Warn().Msg(err.Error())
	}
	err = SendAnswer(w, r, result, err)
	if err != nil {
		logger.Log(r).Warn().Msg(err.Error())
	}
}

// Generic put
func Put(w http.ResponseWriter, r *http.Request) {
	args := FormToMap(r)
	tablename := GetTableName(r)
	data, err := DecodeJSON(r)
	if err == nil {
		err = dbhelper.Update(r, tablename, args, *data)
	}
	if err != nil {
		logger.Log(r).Warn().Msg(err.Error())
	}
	err = SendAnswer(w, r, nil, err)
	if err != nil {
		logger.Log(r).Warn().Msg(err.Error())
	}
}

// Generic delete
func Delete(w http.ResponseWriter, r *http.Request) {
	args := FormToMap(r)
	tablename := GetTableName(r)
	err := dbhelper.Delete(r, tablename, args)
	if err != nil {
		logger.Log(r).Warn().Msg(err.Error())
	}
	err = SendAnswer(w, r, nil, err)
	if err != nil {
		logger.Log(r).Warn().Msg(err.Error())
	}
}
