/*
Package binder provides a simple way for binding an HTTP request body to an struct.

It supports the following media types: JSON, XML, form and multipart form.

Binding is as easy as:

    type MyBody struct {
        Foo string `json:"foo"`
        Bar string `json:"bar"`
    }

    func handler(w http.ResponseWriter, r *http.Request) {
        body := MyBody{}

        err := binder.BindBody(r, &MyBody)
    }

Just give it a pointer to the http.Request and a pointer to the struct to bind.
*/
package binder
