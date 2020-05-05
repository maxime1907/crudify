# Crudify

crudify provides a Golang API to perform crud and custom request on a database

## Makefile

* `libs` : install internal dependencies
* `tests` : run all test files
* `func` : run functional test package
* `bench` : run benchmarks on CRUD requests
* `createdb` : create the schema for test purposes
* `cleandb` : drop the schema for test purposes

## Configuration

Configuration file needed to connect to target database

## Glide

Configuration file to manage dependency versions

## Hierarchy
* cmd : Main entry point of the project (prov/main.go)
* dbhelper : Database Helper package
* handlers : Handler package
* router : Router package (contains a logger)
* tools : Contains every script or file that helped developing this project

## External dependencies

* [Mux](https://github.com/gorilla/mux) : A powerful URL router and dispatcher
* [Dbr](https://github.com/gocraft/dbr/) : Additions to Go's database/sql for super fast performance and convenience
* [Pq](https://godoc.org/github.com/lib/pq) : Postgres driver for database/sql
* [Viper](https://github.com/spf13/viper) : Go configuration with fangs
* [Zerolog](https://github.com/rs/zerolog) : Zero Allocation JSON Logger

## Miscellaneous

* [Glide](https://github.com/Masterminds/glide) : update external dependencies 
* [GoTests](https://github.com/cweill/gotests) : generate test files

## JSON Examples
* Configuration file named `config.json` (Fields below server are optional, interval is in seconds)
```json
{
	"database" : {
		"host" : "127.0.0.1",
		"user" : "test",
		"dbname" : "test",
		"password" : "testpass",
		"sslmode" : "disable",
		"driver" : "postgres"
	},
	"server" : {
		"port" : 8080
	},
	"update" : {
		"url" : "http://test.com:5858/binaries/myapp",
		"interval" : 500
	},
	"tls" : {
		"crt" : "server.crt",
		"key" : "server.key"
	},
	"cors" : {
		"origins" : [
			"*"
		],
		"methods" : [
			"GET",
			"DELETE",
			"POST",
			"PUT",
			"OPTIONS"
		],
		"headers" : [
			"Access-Control-Allow-Origin",
			"Content-Type",
			"X-Requested-With"
		]
	}
}
```