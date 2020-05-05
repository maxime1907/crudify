package router

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/maxime1907/crudify/config"
	"github.com/maxime1907/crudify/dbhelper"
	"github.com/maxime1907/crudify/handler"
	"github.com/maxime1907/crudify/logger"
	"github.com/maxime1907/crudify/auth"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type RouteHelper struct {
	Method string
	Route  string
}

var allRoutes []RouteHelper

func RootGet(w http.ResponseWriter, r *http.Request) {
	err := handler.SendAnswer(w, r, allRoutes, nil)
	if err != nil {
		logger.Log(r).Warn().Msg(err.Error())
	}
}

func PanicFuncHandler(w http.ResponseWriter, r *http.Request, name string, start time.Time) {
	if reco := recover(); reco != nil {
		var err error
		switch x := reco.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("Unknown panic")
		}
		logger.Log(r).Error().Str("stacktrace", identifyPanic()).Msg(err.Error())
		err = handler.SendAnswer(w, r, nil, err)
		if err != nil {
			logger.Log(r).Error().Msg(err.Error())
		}
	}
	logger.Log(r).Info().Msg(r.Method + "\t" + r.RequestURI + "\t" + name + "\t" + time.Since(start).String())
}

// identifyPanic is useful to get runtime stacktrace
// It ignores every function calls after "runtime.panic"
func identifyPanic() string {
	var name, file string
	var pcKeep []uintptr
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for i, p := range pc[:n] {
		fn := runtime.FuncForPC(p)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(p)
		name = fn.Name()

		if !strings.HasPrefix(name, "runtime.") {
			pcKeep = pc[i:]
			break
		}
	}

	var tree string
	for _, p := range pcKeep {
		fn := runtime.FuncForPC(p)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(p)
		name = fn.Name()

		switch {
		case name != "":
			tree += " -> " + fmt.Sprintf("%v:%v", name, line)
		case file != "":
			tree += " -> " + fmt.Sprintf("%v:%v", file, line)
		}
	}
	return tree
}

func AddRoute(router *mux.Router, route Route, routerinfo config.RouterInfo) {
	var handler http.Handler

	handler = route.HandlerFunc

	if (routerinfo.Username != "" && routerinfo.Password != "") {
		handler = auth.BasicAuth(handler, routerinfo)
	}

	handler = logger.Logger(handler, route.Name, PanicFuncHandler)

	router.
		Methods(route.Method).
		Path(route.Pattern).
		Name(route.Name).
		Handler(handler)
}

func AddRoutes(router *mux.Router, routes *[]Route, routerinfo config.RouterInfo) {
	if routes != nil {
		for _, route := range *routes {
			AddRoute(router, route, routerinfo)
			allRoutes = append(allRoutes, RouteHelper{Method: route.Method, Route: route.Pattern})
		}
	}
}

func GetCRUD(gethandler http.HandlerFunc, posthandler http.HandlerFunc,
	puthandler http.HandlerFunc, deletehandler http.HandlerFunc) (*[]Route, error) {
	var routes []Route

	tables, err := dbhelper.GetTables(nil)
	if err != nil {
		return nil, err
	}

	for _, value := range tables {
		routes = append(routes, Route{
			Method:      "GET",
			Pattern:     "/" + value,
			Name:        "get_" + value,
			HandlerFunc: gethandler,
		})
		routes = append(routes, Route{
			Method:      "POST",
			Pattern:     "/" + value,
			Name:        "post_" + value,
			HandlerFunc: posthandler,
		})
		routes = append(routes, Route{
			Method:      "PUT",
			Pattern:     "/" + value,
			Name:        "put_" + value,
			HandlerFunc: puthandler,
		})
		routes = append(routes, Route{
			Method:      "DELETE",
			Pattern:     "/" + value,
			Name:        "delete_" + value,
			HandlerFunc: deletehandler,
		})
	}
	return &routes, nil
}

func GetCORS(router *mux.Router, infos config.CORSInfo) http.Handler {
	originsOk := handlers.AllowedOrigins(infos.Origins)
	methodsOk := handlers.AllowedMethods(infos.Methods)
	headersOk := handlers.AllowedHeaders(infos.Headers)

	handler := handlers.CORS(originsOk, headersOk, methodsOk)(router)

	return handler
}

func NewCustom(custom_routes *[]Route, crud_routes *[]Route, root_get_explicit *Route, routerinfo config.RouterInfo) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	if (custom_routes != nil) {
		AddRoutes(router, custom_routes, routerinfo)
	}

	if (crud_routes != nil) {
		AddRoutes(router, crud_routes, routerinfo)
	}

	if (root_get_explicit != nil) {
		AddRoute(router, *root_get_explicit, routerinfo)
	}

	return router
}

func New(custom_routes *[]Route, enableCRUD bool, enableRootGet bool, routerinfo config.RouterInfo) *mux.Router {
	var crud_routes *[]Route = nil
	var root_get_explicit *Route = nil
	var err error

	router := mux.NewRouter().StrictSlash(true)

	if enableCRUD {
		crud_routes, err = GetCRUD(handler.Get, handler.Post, handler.Put, handler.Delete)
		if err != nil {
			panic(err.Error())
		}
	}

	if enableRootGet {
		root_get_explicit = &Route{
			Method:      "GET",
			Pattern:     "/",
			Name:        "root_get",
			HandlerFunc: RootGet,
		};
	}

	NewCustom(custom_routes, crud_routes, root_get_explicit, routerinfo)

	return router
}

func Run(h http.Handler, port int) error {
	port_s := strconv.Itoa(port)
	logger.Log(nil).Info().Msg("Listening and serving on port " + port_s)
	return http.ListenAndServe(":"+port_s, h)
}

func RunWithTLS(h http.Handler, port int, tlsconf config.TLSInfo) error {
	port_s := strconv.Itoa(port)
	logger.Log(nil).Info().Msg("Listening and serving on port " + port_s + " with TLS")
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	srv := &http.Server{
		Addr:         ":" + port_s,
		Handler:      h,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	return srv.ListenAndServeTLS(tlsconf.Crt, tlsconf.Key)
}

func getPort(l net.Listener) string {
	var myport = "undefined"

	pos := strings.LastIndex(l.Addr().String(), ":")
	if pos > -1 {
		myport = l.Addr().String()[pos+1 : len(l.Addr().String())]
	}
	return myport
}

func RunWithListener(h http.Handler, l net.Listener) error {
	logger.Log(nil).Info().Msg("Serving with custom listener on port " + getPort(l))
	return http.Serve(l, h)
}
