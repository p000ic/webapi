package webapi

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
)

// Serializers - JSON Serializer
var (
	Serializers = map[string]Serializer{
		"application/x-www-form-urlencoded": &formSerializer{},
		"application/json":                  &jsonSerializer{},
		"application/xml":                   &xmlSerializer{},
		"":                                  &jsonSerializer{},
	}
)

type (
	xmlSerializer  struct{}
	jsonSerializer struct{}
	formSerializer struct{}
	responseWriter struct {
		ctx *Context
	}
)

func (*xmlSerializer) Marshal(obj interface{}) ([]byte, error) {
	return xml.Marshal(obj)
}

func (*xmlSerializer) Unmarshal(src []byte, obj interface{}) error {
	return xml.Unmarshal(src, obj)
}

func (*jsonSerializer) Marshal(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func (*jsonSerializer) Unmarshal(src []byte, obj interface{}) error {
	return json.Unmarshal(src, obj)
}

func (*formSerializer) Marshal(obj interface{}) ([]byte, error) {
	src, err := json.Marshal(obj)
	kv := map[string]interface{}{}
	if err == nil {
		err = json.Unmarshal(src, &kv)
	}
	if err != nil {
		return nil, err
	}
	var values = url.Values{}
	for k, v := range kv {
		if t := reflect.TypeOf(v).Kind(); t == reflect.Map || t == reflect.Struct {
			continue
		}
		values.Set(k, fmt.Sprintf("%v", v))
	}
	return []byte(values.Encode()), nil
}

func (*formSerializer) Unmarshal(src []byte, obj interface{}) error {
	val := reflect.ValueOf(obj)
	if val.Type().Kind() == reflect.Struct || !val.Elem().CanSet() {
		return errors.New("type " + val.Type().String() + " is readonly")
	}
	values, err := url.ParseQuery(string(src))
	if err == nil {
		p := &param{
			Type: reflect.TypeOf(obj),
		}
		var value *reflect.Value
		value, err = p.loadFromValues(values)
		if err == nil {
			reflect.ValueOf(obj).Elem().Set(value.Elem())
		}
	}
	return err
}

// Reply Default implementation of Response
type Reply struct {
	Status int
	Body   interface{}
}

// StatusCode HTTP Status Code
func (reply Reply) StatusCode() int {
	return reply.Status
}

// Data Body
func (reply Reply) Data() interface{} {
	return reply.Body
}

func (w *responseWriter) Write(p []byte) (int, error) {
	defer func() {
		if w.ctx.statuscode == 0 {
			w.ctx.statuscode = 200 // mark data has been transferred
		}
	}()
	return w.ctx.w.Write(p)
}

func (w *responseWriter) Header() http.Header {
	return w.ctx.w.Header()
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.ctx.statuscode = statusCode
	w.ctx.w.WriteHeader(statusCode)
}
