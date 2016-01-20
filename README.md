# binder
[![Build Status](https://api.travis-ci.org/marcsv/go-binder.svg?branch=master)](https://travis-ci.org/marcsv/go-binder)
[![GoDoc](http://godoc.org/github.com/marcsv/go-binder/binder?status.svg)](http://godoc.org/github.com/marcsv/go-binder/binder)

Simple package for binding of an HTTP request body easily to an struct.

## Install

Donwload and install the package

    $ go get -u github.com/marcsv/go-binder/binder

## Usage

It supports the following media types: JSON, XML, form and multipart form.

```go
type MyBody struct {
    Foo string `json:"foo"`
    Bar string `json:"bar"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    body := MyBody{}

    err := binder.BindBody(r, &MyBody)
}
```

Just give it a pointer to the http.Request and a pointer to the struct to bind.
