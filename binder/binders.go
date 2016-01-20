package binder

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

const (
	multipartFormMaxMemory = int64(1024 * 1024 * 16) // 16 MB
)

func bindJSON(r *http.Request, target interface{}) error {
	jsonDecoder := json.NewDecoder(r.Body)

	if err := jsonDecoder.Decode(target); err != nil {
		return err
	}

	return nil
}

func bindXML(r *http.Request, target interface{}) error {
	xmlDecoder := xml.NewDecoder(r.Body)

	if err := xmlDecoder.Decode(target); err != nil {
		return err
	}

	return nil
}

func bindForm(r *http.Request, target interface{}) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	return formToStruct(r.Form, nil, target)
}

func bindMultipartForm(r *http.Request, target interface{}) error {
	err := r.ParseMultipartForm(multipartFormMaxMemory)
	if err != nil {
		return err
	}

	return formToStruct(r.MultipartForm.Value, r.MultipartForm.File, target)
}
