package crudify

import (
	"errors"
	"net"
	"net/http"

	"github.com/maxime1907/crudify/config"
	"github.com/maxime1907/crudify/dbhelper"
	"github.com/maxime1907/crudify/router"
)

// Run server with connection to database
func Run(l net.Listener, myconfig *config.Config, routes *[]router.Route, enableCORS bool) error {
	var handler http.Handler

	if myconfig == nil {
		return errors.New("Configuration is not set")
	}
	err := dbhelper.Connect(myconfig.Database)
	if err != nil {
		return err
	}
	myrouter := router.New(routes, true, true, myconfig.Server)
	if enableCORS {
		handler = router.GetCORS(myrouter, myconfig.Cors)
	} else {
		handler = myrouter
	}
	if l != nil {
		return router.RunWithListener(handler, l)
	}
	return router.Run(handler, myconfig.Server.Port)
}
