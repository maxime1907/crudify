package logger

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DeferFunc func(w http.ResponseWriter, r *http.Request, name string, mytime time.Time)

func Pretty() {
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func Log(r *http.Request) *zerolog.Logger {
	var c context.Context

	if r != nil {
		c = r.Context()
	}
	return LogWithContext(c)
}

// LogWithContext takes a context in argument and returns a logger
// We try to search in context if a "uuid" is stored, if found we add it
func LogWithContext(c context.Context) *zerolog.Logger {
	if c == nil {
		return &log.Logger
	}
	if uuid := c.Value("uuid"); uuid != nil {
		mylogger := log.With().Str("UUID", uuid.(string)).Logger()
		return &mylogger
	}
	return &log.Logger
}

// LoggerWithFlags returns a logger with flags
// It can be used to create a logger for a specific package, and use it to automatically log package-specific info like configured back-end
func LogWithFlags(m map[string]string) *zerolog.Logger {
	l := log.With()
	for key, val := range m {
		l.Str(key, val)
	}
	logger := l.Logger()
	return &logger
}

func NewUUID() (string, error) {
	myuuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return myuuid.String(), err
}

func DiffArrays(slice1 []string, slice2 []string) []string {
    var diff []string

    for i := 0; i < 2; i++ {
        for _, s1 := range slice1 {
            found := false
            for _, s2 := range slice2 {
                if s1 == s2 {
                    found = true
                    break
                }
            }
            if !found {
                diff = append(diff, s1)
            }
        }
        if i == 0 {
            slice1, slice2 = slice2, slice1
        }
    }

    return diff
}

func Logger(inner http.Handler, name string, myfunc DeferFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		//Generate uuid and add to parameters
		myuuid, err := NewUUID()
		if err != nil {
			Log(nil).Warn().Msg("Cannot generate a UUID")
		}

		ctx := context.WithValue(r.Context(), "uuid", myuuid)
		myresp := r.WithContext(ctx)

		defer myfunc(w, myresp, name, start)

		inner.ServeHTTP(w, myresp)
	})
}
