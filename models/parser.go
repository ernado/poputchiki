package models

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

const (
	ContentTypeHeader = "Content-Type"
)

func mapToStruct(q url.Values, val interface{}) error {
	nQ := make(map[string]interface{})
	t := reflect.ValueOf(val)
	st := t.Elem().Type()
	fields := st.NumField()
	for i := 0; i < fields; i++ {
		field := st.Field(i)
		key := field.Tag.Get("json")
		if key == "" {
			key = strings.ToLower(field.Name)
		}
		if strings.Index(key, ",") != -1 {
			key = strings.Split(key, ",")[0]
		}
		if q.Get(key) == "" {
			continue
		}
		value := q[key]
		if len(value) == 1 {
			v := value[0]
			vInt, err := strconv.Atoi(v)
			if err != nil || field.Type.Name() == "string" {
				nQ[key] = v
			} else {
				nQ[key] = vInt
			}
		} else {
			nQ[key] = value
		}
	}
	j, err := json.Marshal(nQ)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, val)

}

func Parse(r *http.Request, v interface{}) error {
	contentType := r.Header.Get(ContentTypeHeader)
	if strings.Index(contentType, "json") != -1 {
		decoder := json.NewDecoder(r.Body)
		return decoder.Decode(v)
	}
	if strings.Index(contentType, "form") != -1 {
		err := r.ParseForm()
		if err != nil {
			log.Println("parse frm", err)
		}
		if len(r.Form) > 0 {
			return mapToStruct(r.Form, v)
		}
		if len(r.PostForm) > 0 {
			return mapToStruct(r.PostForm, v)
		}
	}

	return mapToStruct(r.URL.Query(), v)
}

type Parser struct {
	r *http.Request
}

func (p Parser) Parse(v interface{}) error {
	return Parse(p.r, v)
}

func NewParser(r *http.Request) Parser {
	return Parser{r}
}
