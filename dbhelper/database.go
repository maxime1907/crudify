package dbhelper

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gocraft/dbr"
	_ "github.com/lib/pq"
	"github.com/maxime1907/crudify/config"
	"github.com/maxime1907/crudify/logger"
)

// Global variable that holds all table names in database
var tables map[string]string

// Global variable that holds connection to database
var connection *dbr.Connection

var REQUEST_ARG_PREFIX string = "_"

type Builder struct {
	Column string
	Value string
	Operand string
}

func GetConnection() *dbr.Connection {
	return connection
}

// Get tables from sql database
func GetTables(r *http.Request) (map[string]string, error) {
	logger.Log(r).Debug().Msg("Getting table names from database")
	if tables == nil {
		args := map[string]string{"table_schema": `public`, "table_type": `BASE TABLE`}
		res, err := Select(r, "information_schema.tables", args)
		if err != nil {
			return nil, err
		}
		size := len(*res)
		if size > 0 {
			tables = map[string]string{}
			for i := 0; i < size; i++ {
				tables[(*res)[i]["table_name"].(string)] = (*res)[i]["table_name"].(string)
			}
		} else {
			return nil, errors.New("Database does not contain any table OR you do not have proper rights to access it")
		}
	}
	return tables, nil
}

// Convert DBInfo into string
func FormatSettings(dbinfo config.DBInfo) string {
	logger.Log(nil).Debug().Msg("Formatting connection parameters")
	var dbSourceName string
	val := reflect.Indirect(reflect.ValueOf(dbinfo))
	for i := 0; i < val.NumField(); i++ {
		fieldName := strings.ToLower(val.Type().Field(i).Name)
		fieldValue := val.Field(i).String()
		if fieldName != "driver" {
			dbSourceName += fieldName + "=" + fieldValue + " "
		}
	}
	if val.NumField() >= 1 {
		sz := len(dbSourceName)
		dbSourceName = dbSourceName[:sz-1]
	}
	return dbSourceName
}

// Read config file, open and ping database
func Connect(configuration config.DBInfo) error {
	logger.Log(nil).Debug().Msg("Connecting to database")
	var err error

	dsn := FormatSettings(configuration)
	logger.Log(nil).Debug().Msg("Opening connection to " + configuration.Host + " with SSL " + configuration.Sslmode + "d")
	connection, err = dbr.Open(configuration.Driver, dsn, nil)
	if err != nil {
		return err
	}
	logger.Log(nil).Debug().Msg("Pinging...")
	err = connection.DB.Ping()
	if err != nil {
		return err
	}
	return nil
}

// Exec query and returns result into json
func ExecQueryJSON(r *http.Request, query string) (*[]map[string]interface{}, error) {
	logger.Log(r).Debug().Msg("Executing on database query => " + query)
	var result []map[string]interface{}
	var rows *sql.Rows
	var err error

	rows, err = connection.DB.Query(query)
	if err != nil {
		return nil, err
	}

	logger.Log(r).Debug().Msg("Mapping result query")
	var cols []string
	cols, err = rows.Columns()
	if err != nil {
		return nil, err
	}

	// Result is your slice string.
	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	var finalrows map[string]interface{}
	var raw []byte
	var i int

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			return nil, err
		}

		finalrows = make(map[string]interface{})

		for i, raw = range rawResult {
			if raw == nil {
				finalrows[cols[i]] = nil
			} else {
				finalrows[cols[i]] = string(raw)
			}
		}
		result = append(result, finalrows)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Select primary keys of given table
func SelectPrimaryKeys(r *http.Request, tablename string) (*[]map[string]interface{}, error) {
	logger.Log(r).Debug().Msg("Selecting primary keys on table: " + tablename)

	var myselect = []string{
		"pg_attribute.attname",
		"format_type(pg_attribute.atttypid, pg_attribute.atttypmod)",
	}

	var from = "pg_index, pg_class, pg_attribute, pg_namespace"

	var where = map[string]string{
		"pg_class.oid":          "^" + tablename + "^::regclass",
		"indrelid":              "\"pg_class\".\"oid\"",
		"nspname":               "^public^",
		"pg_class.relnamespace": "\"pg_namespace\".\"oid\"",
		"pg_attribute.attrelid": "\"pg_class\".\"oid\"",
		"pg_attribute.attnum":   "any(\"pg_index\".\"indkey\")",
		"indisprimary":          "^true^",
	}

	query, err := SelectByQueryArgs(myselect, from, where)
	if err != nil {
		return nil, err
	}

	query = strings.Replace(query, "'", "", -1)
	query = strings.Replace(query, "^", "'", -1)

	return ExecQueryJSON(r, query)
}

// Select primary keys of given table
func SelectForeignKeys(r *http.Request, tablename string) (*[]map[string]interface{}, error) {
	logger.Log(r).Debug().Msg("Selecting foreign keys on table: " + tablename)

	var myselect = []string{
		"kcu.column_name",
		"ccu.table_schema AS foreign_table_schema",
		"ccu.table_name AS foreign_table_name",
		"ccu.column_name AS foreign_column_name",
	}

	var from = "information_schema.table_constraints AS tc " +
		"JOIN information_schema.key_column_usage AS kcu " +
		"ON tc.constraint_name = kcu.constraint_name " +
		"AND tc.table_schema = kcu.table_schema " +
		"JOIN information_schema.constraint_column_usage AS ccu " +
		"ON ccu.constraint_name = tc.constraint_name " +
		"AND ccu.table_schema = tc.table_schema"

	var where = []Builder{
		Builder{"tc.constraint_type", "FOREIGN KEY", "="},
		Builder{"tc.table_name", tablename, "="},
	}
	return SelectWithQuery(r, myselect, from, map[string]string{}, where)
}

func SelectWithQuery(r *http.Request, myselect []string, from string, args map[string]string, where []Builder) (*[]map[string]interface{}, error) {
	var result *[]map[string]interface{}

	logger.Log(r).Debug().Msg("Selecting on table: " + from)

	query, err := SelectByQuery(myselect, from, args, where)
	if err != nil {
		return nil, err
	}
	result, err = ExecQueryJSON(r, query)
	if (err == nil) {
		if _, ok := args[REQUEST_ARG_PREFIX + "nested"]; ok {
			result, err = AddNestedObjects(r, result, from)
		}
	}
	return result, err
}

// Convert SelectStmt to sql query
func builderToQuery(builder dbr.Builder, d dbr.Dialect, values []interface{}) (string, error) {
	var query string
	var err error

	buf := dbr.NewBuffer()
	err = builder.Build(d, buf)
	if err != nil {
		return query, err
	}
	query, err = dbr.InterpolateForDialect(buf.String(), values, d)
	if err != nil {
		return query, err
	}
	return query, nil
}

func GetArgValues(args map[string]string) []interface{} {
	var values []interface{}
	for key, value := range args {
		if !strings.HasPrefix(key, REQUEST_ARG_PREFIX) {
			values = append(values, value)
		}
	}
	return values
}

func GetBuilderValues(where []Builder) []interface{} {
	var values []interface{}
	for _, build := range where {
		values = append(values, build.Value)
	}
	return values
}

func SelectByQueryArgs(from []string, tablename string, args map[string]string) (string, error) {
	return SelectByQuery(from, tablename, args, []Builder{})
}

func SelectByQuery(from []string, tablename string, args map[string]string, where []Builder) (string, error) {
	var values []interface{} = make([]interface{}, 0)
	var query string
	var err error

	if GetConnection() == nil {
		return "", errors.New("Not connected to database")
	}
	dbrSess := connection.NewSession(nil)

	//Build our query
	builder := dbrSess.Select(from...)

	builder, err = AddArgs(builder, tablename, args, &values)
	if (err != nil) {
		return "", err
	}

	if (len(where) > 0) {
		builder, err = AddWhere(builder, where, &values)
		if (err != nil) {
			return "", err
		}
	}

	query, err = builderToQuery(builder, connection.Dialect, values)
	if err != nil {
		return "", err
	}
	return query, nil
}

func ChooseWhereStatement(build Builder) (dbr.Builder, error) {
	switch (build.Operand) {
	case "<=":
		return dbr.Lte(build.Column, build.Value), nil
	case "<":
		return dbr.Lt(build.Column, build.Value), nil
	case ">":
		return dbr.Gt(build.Column, build.Value), nil
	case ">=":
		return dbr.Gte(build.Column, build.Value), nil
	case "=":
		return dbr.Eq(build.Column, build.Value), nil
	}
	return nil, errors.New("Where statement not recognised: " + build.Operand)
}

func AddWhere(builder *dbr.SelectStmt, where []Builder, values *[]interface{}) (*dbr.SelectStmt, error) {
	for _, build := range where {
		statement, err := ChooseWhereStatement(build)
		if (err == nil) {
			builder.Where(statement)
			*values = append(*values, build.Value)
		} else {
			return builder, err
		}
	}
	return builder, nil
}

func AddArgs(builder *dbr.SelectStmt, tablename string, args map[string]string, values *[]interface{}) (*dbr.SelectStmt,  error) {
	if _, ok := args[REQUEST_ARG_PREFIX + "only"]; ok {
		tablename = "ONLY " + tablename
	}

	builder = builder.From(tablename)

	for key, value := range args {
		if !strings.HasPrefix(key, REQUEST_ARG_PREFIX) {
			builder = builder.Where(dbr.Eq(key, value))
			*values = append(*values, value)
		}
	}

	if val, ok := args[REQUEST_ARG_PREFIX + "orderby"]; ok {
		if val2, ok2 := args[REQUEST_ARG_PREFIX + "order"]; ok2 {
			switch val2 {
			case "true":
				builder = builder.OrderAsc(val)
			case "false":
				builder = builder.OrderDesc(val)
			default:
				return nil, errors.New("Order should be true or false")
			}
		} else {
			builder = builder.OrderAsc(val)
		}
	}

	if val, ok := args[REQUEST_ARG_PREFIX + "limit"]; ok {
		var ret uint64
		ret, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, errors.New("Limit value \"" + val + "\" is not a valid number")
		}
		builder = builder.Limit(ret)
	}
	return builder, nil
}

func AddNestedObjects(r *http.Request, results *[]map[string]interface{}, tablename string) (*[]map[string]interface{}, error) {
	var resultsNested *[]map[string]interface{}
	logger.Log(r).Debug().Msg("Adding nested objects on table: " + tablename)

	if (results == nil || len(*results) <= 0) {
		return results, nil
	}
	foreignKeys, err := SelectForeignKeys(r, tablename)
	if (err != nil) {
		return results, err
	}
	if (foreignKeys == nil || len(*foreignKeys) <= 0) {
		return results, nil
	}
	for _, resultMap := range *results {
		for _, foreignKeyMap := range *foreignKeys {
			foreign_table_name := fmt.Sprintf("%v", foreignKeyMap["foreign_table_name"])
			foreign_column_name := fmt.Sprintf("%v", foreignKeyMap["foreign_column_name"])
			column_name := fmt.Sprintf("%v", foreignKeyMap["column_name"])
			
			args := map[string]string{
				foreign_column_name : fmt.Sprintf("%v", resultMap[column_name]),
				REQUEST_ARG_PREFIX + "nested" : "",
			}

			resultsNested, err = Select(r, foreign_table_name, args)
			if (err != nil) {
				return results, err
			} else {
				if (len(*resultsNested) == 1) {
					resultMap[foreign_table_name + "_obj"] = *resultsNested
				}
			}
		}
	}
	return results, nil
}

// Select retrieves row(s)
func Select(r *http.Request, tablename string, args map[string]string) (*[]map[string]interface{}, error) {
	return SelectWithQuery(r, []string{"*"}, tablename, args, []Builder{})
}

// Insert add row(s)
func Insert(r *http.Request, tablename string, args map[string]string, json []map[string]interface{}) (*[]map[string]interface{}, error) {
	logger.Log(r).Debug().Msg("Inserting on table: " + tablename)

	var builder *dbr.InsertStmt
	var id int64 = 0
	var size_keys int
	var values []interface{}
	var keys []string
	var k string
	var v interface{}

	if GetConnection() == nil {
		return nil, errors.New("Not connected to database")
	}
	dbrSess := connection.NewSession(nil)

	size_json := len(json)
	if size_json <= 0 {
		return nil, errors.New("Missing data in json")
	}

	tx, err := dbrSess.Begin()
	if err != nil {
		return nil, err
	}

	val, returning := args[REQUEST_ARG_PREFIX + "returning"];

	defer tx.RollbackUnlessCommitted()

	for i := 0; i < size_json; i++ {
		//Build our query
		builder = tx.InsertInto(tablename)

		size_keys = len(json[i])
		keys = make([]string, 0, size_keys)
		values = make([]interface{}, 0, size_keys)

		for k, v = range json[i] {
			keys = append(keys, k)
			values = append(values, v)
		}

		builder.Value = append(builder.Value, values)
		builder.Column = keys

		if returning {
			builder = builder.Returning(val)
			err = builder.Load(&id)
		} else {
			_, err = builder.Exec()
		}

		if err != nil {
			return nil, err
		}

		if returning && id > 0 {
			json[i]["id"] = id
		}
	}
	if returning {
		return &json, tx.Commit()
	}
	return nil, tx.Commit()
}

// Update upgrade row(s)
func Update(r *http.Request, tablename string, args map[string]string, json []map[string]interface{}) error {
	logger.Log(r).Debug().Msg("Updating on table: " + tablename)

	var builder *dbr.UpdateStmt
	var value interface{}
	var key string
	var v map[string]interface{}
	var result sql.Result
	var nb int64

	if GetConnection() == nil {
		return errors.New("Not connected to database")
	}
	dbrSess := connection.NewSession(nil)

	size_json := len(json)
	if size_json <= 0 {
		return errors.New("Missing data in json")
	}

	res, err := SelectPrimaryKeys(r, tablename)
	if err != nil {
		return err
	}

	tx, err := dbrSess.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	size_res := len(*res)

	for i := 0; i < size_json; i++ {
		builder = tx.Update(tablename)

		var myvalues []interface{}
		for key, value = range json[i] {
			builder = builder.Set(key, value)
			myvalues = append(myvalues, value)
		}

		var pk_fields_check []string
		var pk_fields []string
		for key, value = range json[i] {
			for _, v = range *res {
				pk_fields = append(pk_fields, fmt.Sprintf("%v", v["attname"]))
				if key == v["attname"] {
					pk_fields_check = append(pk_fields_check, key)
					builder = builder.Where(dbr.Eq(key, value))
				}
			}
		}
		if len(pk_fields_check) != size_res {
			return errors.New("Missing primary keys in json (" + 
				strings.Join(logger.DiffArrays(pk_fields_check, pk_fields), ", ") + ")")
		}

		result, err = builder.Exec()
		if err == nil {
			nb, err = result.RowsAffected()
		}

		if err != nil {
			return err
		} else if nb <= 0 {
			return errors.New("sql: no rows in result set for " + fmt.Sprintf("%#v", json[i]))
		}
	}
	return tx.Commit()
}

// Delete removes row(s)
func Delete(r *http.Request, tablename string, args map[string]string) error {
	logger.Log(r).Debug().Msg("Deleting on table: " + tablename)
	if GetConnection() == nil {
		return errors.New("Not connected to database")
	}
	if !(args != nil && len(args) > 0) {
		return errors.New("Delete on all rows is disabled")
	}

	dbrSess := connection.NewSession(nil)

	//Build our query
	builder := dbrSess.DeleteFrom(tablename)

	for key, value := range args {
		builder = builder.Where(dbr.Eq(key, value))
	}

	_, err := builder.Exec()
	return err
}

// Delete removes multiple row(s)
func DeleteMultiple(r *http.Request, tablename string, args []map[string]interface{}) error {
	logger.Log(r).Debug().Msg("Deleting multiple rows on table: " + tablename)

	var builder *dbr.DeleteStmt
	var value interface{}
	var key string

	if GetConnection() == nil {
		return errors.New("Not connected to database")
	}

	size := len(args)
	if size <= 0 {
		return errors.New("Missing data in arguments")
	}

	dbrSess := connection.NewSession(nil)

	tx, err := dbrSess.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	for i := 0; i < size; i++ {
		if len(args[i]) > 0 {
			//Build our query
			builder = tx.DeleteFrom(tablename)

			for key, value = range args[i] {
				builder = builder.Where(dbr.Eq(key, value))
			}

			_, err = builder.Exec()
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
