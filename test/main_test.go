package test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/maxime1907/crudify"
	"github.com/maxime1907/crudify/config"
	"github.com/maxime1907/crudify/logger"
)

var url string
var myconfig config.Config

func initUrl() error {
	var err error

	myconfig, err = config.Read("test_config", "../tools/")
	if err != nil {
		return errors.New("Cannot read database configuration file")
	}
	url = "http://" + myconfig.Database.Host + ":" + strconv.Itoa(myconfig.Server.Port) + "/"
	return nil
}

func execRequest(req *http.Request) error {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	mybody, _ := ioutil.ReadAll(resp.Body)
	body := string(mybody)

	if !(resp.Status == "200 OK" && resp.Header["Content-Type"][0] == "application/json; charset=UTF-8" && len(body) > 0) {
		return errors.New("Expected answer was 200 OK but got " + resp.Status)
	}

	var data interface{}
	err = json.Unmarshal(mybody, &data)
	if err != nil {
		return err
	}
	switch val := data.(type) {
	case map[string]interface{}:
		switch valField := val["message"].(type) {
		case string:
			if valField != "" {
				return errors.New("message field problem: " + valField)
			}
		default:
			return errors.New("message field")
		}
	default:
		return errors.New("Wrong json field type")
	}
	return nil
}

func waitRouter(t *testing.T) {
	var start = time.Now()

	for {
		if time.Now().Sub(start).Seconds() >= 10 {
			t.Errorf("Timeout connecting to server")
		}
		time.Sleep(time.Second)
		logger.Log(nil).Debug().Msg("Trying to connect to " + url)
		resp, err := http.Get(url)
		if err != nil {
			logger.Log(nil).Debug().Msg("Failed")
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logger.Log(nil).Debug().Msg("Status code is not 200 OK")
			continue
		}

		// Reached this point: server is up and running!
		break
	}
}

func launchTests(t *testing.T) {
	waitRouter(t)
	testPostSingle(t)
	testPostMultiple(t)
	testPutSingle(t)
	testPutMultiple(t)
	testRoot(t)
	testGetSingle(t)
	testGetMultiple(t)
	testDelete(t)
	testDeleteAllTest(t)

}

func TestMain(t *testing.T) {
	err := initUrl()
	if err != nil {
		t.Fatal(err)
	}

	go crudify.Run(nil, &myconfig, nil, false)

	launchTests(t)
}
