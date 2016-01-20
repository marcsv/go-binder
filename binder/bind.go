package binder

import (
	"errors"
	"net/http"
	"strings"
)

const (
	jsonType      = "json"
	xmlType       = "xml"
	formType      = "form-urlencoded"
	multipartType = "multipart/form-data"
)

var (
	errMissingContentType     = errors.New("missing Content-Type")
	errUnsupportedContentType = errors.New("unsupported Content-Type")
)

// BindBody binds the body content of the provided HTTP request into the given
// struct.
func BindBody(r *http.Request, target interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return errMissingContentType
	}

	if strings.Contains(contentType, jsonType) {
		return bindJSON(r, target)
	} else if strings.Contains(contentType, xmlType) {
		return bindXML(r, target)
	} else if strings.Contains(contentType, formType) {
		return bindForm(r, target)
	} else if strings.Contains(contentType, multipartType) {
		return bindMultipartForm(r, target)
	}

	return errUnsupportedContentType
}
