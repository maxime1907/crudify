package benchmark

import (
	"errors"
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

type Test struct {
	Name     string
	Quantity int
	Func     func(int) (time.Duration, error)
}

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

	if resp.Status != "200 OK" {
		return errors.New("Expected status code 200 OK, but got " + resp.Status)
	}
	return nil
}

func execTests(t *testing.T, tests []Test) {
	var period time.Duration
	var err error

	waitRouter(t)
	for i := 0; i < len(tests); i++ {
		period, err = tests[i].Func(tests[i].Quantity)
		if err != nil {
			t.Fatal(tests[i].Name + ": " + err.Error())
		}
		t.Log(tests[i].Name, "with", tests[i].Quantity, "in quantity took", period)
	}
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

func TestMain(t *testing.T) {
	var tests = []Test{
		Test{Name: "Get", Quantity: 1000, Func: testGet},
		Test{Name: "Post", Quantity: 1000, Func: testPost},
		Test{Name: "Put", Quantity: 1000, Func: testPut},
		Test{Name: "Delete", Quantity: 1000, Func: testDelete},
	}

	err := initUrl()
	if err != nil {
		t.Fatal(err)
	}

	go crudify.Run(nil, &myconfig, nil, false)

	execTests(t, tests)
}
