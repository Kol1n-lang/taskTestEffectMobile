package utils

import (
	"github.com/gorilla/schema"
	"net/http"
	"reflect"
	"time"
)

func QueryParser(r *http.Request, dest interface{}) error {
	decoder := schema.NewDecoder()
	decoder.RegisterConverter(time.Time{}, func(s string) reflect.Value {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return reflect.Value{}
		}
		return reflect.ValueOf(t)
	})
	return decoder.Decode(dest, r.URL.Query())
}
