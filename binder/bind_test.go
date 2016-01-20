package binder

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type Body struct {
	Foo string `json:"foo" xml:"foo"`
	Bar string `json:"bar" xml:"bar"`
}

type FilesBody struct {
	File      *multipart.FileHeader   `form:"file"`
	FileSlice []*multipart.FileHeader `form:"file_slice"`
}

type Unknown struct{}

type Embedded struct {
	Foobar string `form:"foobar"`
}

type FormBody struct {
	String   string  `form:"string"`
	Bool     bool    `form:"bool"`
	Int      int     `form:"int"`
	Int8     int8    `form:"int8"`
	Int16    int16   `form:"int16"`
	Int32    int32   `form:"int32"`
	Int64    int64   `form:"int64"`
	Uint     uint    `form:"uint"`
	Uint8    uint8   `form:"uint8"`
	Uint16   uint16  `form:"uint16"`
	Uint32   uint32  `form:"uint32"`
	Uint64   uint64  `form:"uint64"`
	Float32  float32 `form:"float32"`
	Float64  float64 `form:"float64"`
	IntSlice []int   `form:"int_slice"`
	Embedded
	NoTag   string
	Unknown Unknown `form:"unknown"`
}

type Executable interface {
	execute(t *testing.T)
}

type testCase struct {
	description string
	contentType string
	body        io.Reader
	expectedErr error
}

type standardTestCase struct {
	testCase
	expectedBody Body
}

func (c standardTestCase) execute(t *testing.T) {
	target := Body{}

	r := newRequest(c.contentType, c.body)

	err := BindBody(r, &target)

	assertError(t, c.description, err, c.expectedErr)
	assertBody(t, c.description, target, c.expectedBody)
}

type fileData struct {
	name    string
	content string
}

type multipartTestCase struct {
	testCase
	fileData      *fileData
	fileSliceData []*fileData
}

func (c multipartTestCase) execute(t *testing.T) {
	target := FilesBody{}

	r := newMultipartRequest(c)

	err := BindBody(r, &target)

	assertError(t, c.description, err, c.expectedErr)
	assertSingleFile(t, c.description, target.File, c.fileData)
	assertFilesSlice(t, c.description, target.FileSlice, c.fileSliceData)
}

type formTestCase struct {
	testCase
	inAndOutBody FormBody
}

func (c formTestCase) execute(t *testing.T) {
	target := FormBody{}

	r := newRequest(c.contentType, c.body)

	err := BindBody(r, &target)

	assertError(t, c.description, err, c.expectedErr)
	assertFormBody(t, c.description, target, c.inAndOutBody)
}

var cases = []Executable{
	standardTestCase{
		testCase: testCase{
			description: "Test missing Content-Type error",
			contentType: "",
			body:        strings.NewReader(""),
			expectedErr: errMissingContentType,
		},
		expectedBody: Body{},
	},
	standardTestCase{
		testCase: testCase{
			description: "Test unsupported Content-Type error",
			contentType: "unsupported",
			body:        strings.NewReader(""),
			expectedErr: errUnsupportedContentType,
		},
		expectedBody: Body{},
	},
	standardTestCase{
		testCase: testCase{
			description: "Test JSON binding success",
			contentType: "application/json",
			body:        strings.NewReader(`{"foo":"foo","bar":"bar"}`),
			expectedErr: nil,
		},
		expectedBody: Body{
			Foo: "foo",
			Bar: "bar",
		},
	},
	standardTestCase{
		testCase: testCase{
			description: "Test JSON binding decode error",
			contentType: "application/json",
			body:        strings.NewReader(`{"foo":"foo","bar":"bar"`),
			expectedErr: errors.New("unexpected EOF"),
		},
		expectedBody: Body{},
	},
	standardTestCase{
		testCase: testCase{
			description: "Test XML binding success",
			contentType: "application/xml",
			body:        strings.NewReader(`<body><foo>foo</foo><bar>bar</bar></body>`),
			expectedErr: nil,
		},
		expectedBody: Body{
			Foo: "foo",
			Bar: "bar",
		},
	},
	standardTestCase{
		testCase: testCase{
			description: "Test XML binding decode error",
			contentType: "application/json",
			body:        strings.NewReader(`<body><foo>foo</foo><bar>bar</bar></body`),
			expectedErr: errors.New("invalid character '<' looking for beginning of value"),
		},
		expectedBody: Body{},
	},
	multipartTestCase{
		testCase: testCase{
			description: "Test multipart with single file success",
			contentType: "",
			body:        nil,
			expectedErr: nil,
		},
		fileData: &fileData{
			name:    "file.txt",
			content: "file text content",
		},
		fileSliceData: nil,
	},
	multipartTestCase{
		testCase: testCase{
			description: "Test multipart with multiple files success",
			contentType: "",
			body:        nil,
			expectedErr: nil,
		},
		fileData: nil,
		fileSliceData: []*fileData{
			{
				name:    "file1.txt",
				content: "file text content 1",
			}, {
				name:    "file2.txt",
				content: "file text content 2",
			},
		},
	},
	formTestCase{
		testCase: testCase{
			description: "Test form binding success: string, bool and floats",
			contentType: "application/x-www-form-urlencoded",
			body:        strings.NewReader("string=foo&bool=true&float32=1.9&float64=2.5"),
			expectedErr: nil,
		},
		inAndOutBody: FormBody{
			String:  "foo",
			Bool:    true,
			Float32: 1.9,
			Float64: 2.5,
		},
	},
	formTestCase{
		testCase: testCase{
			description: "Test form binding success: ints",
			contentType: "application/x-www-form-urlencoded",
			body:        strings.NewReader("int=1&int8=2&int16=3&int32=4&int64=5"),
			expectedErr: nil,
		},
		inAndOutBody: FormBody{
			Int:   1,
			Int8:  2,
			Int16: 3,
			Int32: 4,
			Int64: 5,
		},
	},
	formTestCase{
		testCase: testCase{
			description: "Test form binding success: uints",
			contentType: "application/x-www-form-urlencoded",
			body:        strings.NewReader("uint=1&uint8=2&uint16=3&uint32=4&uint64=5"),
			expectedErr: nil,
		},
		inAndOutBody: FormBody{
			Uint:   1,
			Uint8:  2,
			Uint16: 3,
			Uint32: 4,
			Uint64: 5,
		},
	},
	formTestCase{
		testCase: testCase{
			description: "Test form binding success: slice",
			contentType: "application/x-www-form-urlencoded",
			body:        strings.NewReader("int_slice=1&int_slice=2&int_slice=3&int_slice=4"),
			expectedErr: nil,
		},
		inAndOutBody: FormBody{
			IntSlice: []int{1, 2, 3, 4},
		},
	},
	formTestCase{
		testCase: testCase{
			description: "Test form binding success: embedded",
			contentType: "application/x-www-form-urlencoded",
			body:        strings.NewReader("foobar=foobar"),
			expectedErr: nil,
		},
		inAndOutBody: FormBody{
			Embedded: Embedded{
				Foobar: "foobar",
			},
		},
	},
	formTestCase{
		testCase: testCase{
			description: "Test form binding unknown field type error",
			contentType: "application/x-www-form-urlencoded",
			body:        strings.NewReader("unknown=unknown"),
			expectedErr: errUnknownFieldType,
		},
		inAndOutBody: FormBody{},
	},
	formTestCase{
		testCase: testCase{
			description: "Test form binding all fields empty",
			contentType: "application/x-www-form-urlencoded",
			body:        strings.NewReader("string=&bool=&int=&in8=&int16=&int32=&int64=uint=&uin8=&uint16=&uint32=&uint64=&float32=&float64=&int_slice="),
			expectedErr: nil,
		},
		inAndOutBody: FormBody{},
	},
	formTestCase{
		testCase: testCase{
			description: "Test form binding slice type error",
			contentType: "application/x-www-form-urlencoded",
			body:        strings.NewReader("int_slice=foo"),
			expectedErr: errors.New(`strconv.ParseInt: parsing "foo": invalid syntax`),
		},
		inAndOutBody: FormBody{},
	},
	formTestCase{
		testCase: testCase{
			description: "Test form binding parse error",
			contentType: "application/x-www-form-urlencoded",
			body:        nil,
			expectedErr: errors.New("missing form body"),
		},
		inAndOutBody: FormBody{},
	},
}

func TestBindBody(t *testing.T) {
	for _, c := range cases {
		c.execute(t)
	}
}

func newRequest(contentType string, body io.Reader) *http.Request {
	r, err := http.NewRequest("POST", "/", body)
	if err != nil {
		panic("error building request. Error:" + err.Error())
	}

	r.Header.Set("Content-Type", contentType)

	return r
}

func newMultipartRequest(c multipartTestCase) *http.Request {
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)
	defer w.Close()

	if c.fileData != nil {
		formFile, err := w.CreateFormFile("file", c.fileData.name)
		if err != nil {
			panic("error creating form file. Error:" + err.Error())
		}

		formFile.Write([]byte(c.fileData.content))
	}

	if c.fileSliceData != nil {
		for _, file := range c.fileSliceData {
			formFileSlice, err := w.CreateFormFile("file_slice", file.name)
			if err != nil {
				panic("error creating form file_slice. Error:" + err.Error())
			}

			formFileSlice.Write([]byte(file.content))
		}
	}

	return newRequest(w.FormDataContentType(), b)
}

func assertError(t *testing.T, description string, err, expectedErr error) {
	if err != nil && expectedErr != nil {
		if !reflect.DeepEqual(err.Error(), expectedErr.Error()) {
			t.Error("wrong error in:", description, "\nExpected:", expectedErr, "\nGot:", err)
		}
	}
}

func assertBody(t *testing.T, description string, body, expectedBody Body) {
	if !reflect.DeepEqual(body, expectedBody) {
		t.Error("wrong body in:", description, "\nExpected:", expectedBody, "\nGot:", body)
	}
}

func assertFormBody(t *testing.T, description string, body, expectedBody FormBody) {
	if !reflect.DeepEqual(body, expectedBody) {
		t.Error("wrong form body in:", description, "\nExpected:", expectedBody, "\nGot:", body)
	}
}

func assertSingleFile(t *testing.T, description string, fh *multipart.FileHeader, fileData *fileData) {
	if fileData == nil {
		return
	}

	if fh == nil {
		t.Error("wrong FileHeader in:", description, "\nGot: nil\nExpected:", fileData)
	}

	fhContent := fileHeaderContent(fh)

	if fhContent != fileData.content {
		t.Error("wrong filename in:", description, "\nExpected:", fileData.content, "\nGot:", fhContent)
	}

	if fh.Filename != fileData.name {
		t.Error("wrong filename in:", description, "\nExpected:", fileData.name, "\nGot:", fh.Filename)
	}
}

func assertFilesSlice(t *testing.T, description string, fhSlice []*multipart.FileHeader, filesData []*fileData) {
	if filesData == nil {
		return
	}

	if fhSlice == nil {
		t.Error("wrong FileHeader slice in:", description, "\nGot: nil\nExpected:", filesData)
	}

	if len(fhSlice) != len(filesData) {
		t.Error("wrong FileHeader slice count in:", description, "\nGot:", len(fhSlice), "\nExpected:", len(filesData))
	}

	for i, fh := range fhSlice {
		assertSingleFile(t, description, fh, filesData[i])
	}
}

func fileHeaderContent(fh *multipart.FileHeader) string {
	if fh == nil {
		return ""
	}

	f, err := fh.Open()
	defer f.Close()
	if err != nil {
		panic("error opening FileHeader. Error:" + err.Error())
	}

	var b bytes.Buffer
	_, err = b.ReadFrom(f)
	if err != nil {
		panic("error reading from FileHeader. Error:" + err.Error())
	}

	return b.String()
}
