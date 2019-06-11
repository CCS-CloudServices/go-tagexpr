package binding_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/bytedance/go-tagexpr/binding"
	"github.com/henrylee2cn/goutil/httpbody"
	"github.com/stretchr/testify/assert"
)

func TestRawBody(t *testing.T) {
	type Recv struct {
		rawBody **struct {
			A []byte   `rawbody:""`
			B *[]byte  `rawbody:",required"`
			C **[]byte `rawbody:"required"`
			D string   `rawbody:""`
			E *string  `rawbody:""`
			F **string `rawbody:"" vd:"@:len($)<3; msg:'too long'"`
		}
		S string `rawbody:""`
	}
	bodyBytes := []byte("rawbody.............")
	req := newRequest("", nil, nil, bytes.NewReader(bodyBytes))
	recv := new(Recv)
	binder := binding.New(nil)
	err := binder.BindAndValidate(recv, req, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "too long", err.Error())
	for _, v := range []interface{}{
		(**recv.rawBody).A,
		*(**recv.rawBody).B,
		**(**recv.rawBody).C,
		[]byte((**recv.rawBody).D),
		[]byte(*(**recv.rawBody).E),
		[]byte(**(**recv.rawBody).F),
		[]byte(recv.S),
	} {
		assert.Equal(t, bodyBytes, v)
	}
}

func TestQueryString(t *testing.T) {
	type Recv struct {
		X **struct {
			A []string  `query:"a"`
			B string    `query:"b"`
			C *[]string `query:"c,required"`
			D *string   `query:"d"`
		}
		Y string  `query:"y,required"`
		Z *string `query:"z"`
	}
	req := newRequest("http://localhost:8080/?a=a1&a=a2&b=b1&c=c1&c=c2&d=d1&d=d2&y=y1", nil, nil, nil)
	recv := new(Recv)
	binder := binding.New(nil)
	err := binder.BindAndValidate(recv, req, nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a1", "a2"}, (**recv.X).A)
	assert.Equal(t, "b1", (**recv.X).B)
	assert.Equal(t, []string{"c1", "c2"}, *(**recv.X).C)
	assert.Equal(t, "d1", *(**recv.X).D)
	assert.Equal(t, "y1", recv.Y)
	assert.Equal(t, (*string)(nil), recv.Z)
}

func TestQueryNum(t *testing.T) {
	type Recv struct {
		X **struct {
			A []int     `query:"a"`
			B int32     `query:"b"`
			C *[]uint16 `query:"c,required"`
			D *float32  `query:"d"`
		}
		Y bool   `query:"y,required"`
		Z *int64 `query:"z"`
	}
	req := newRequest("http://localhost:8080/?a=11&a=12&b=21&c=31&c=32&d=41&d=42&y=true", nil, nil, nil)
	recv := new(Recv)
	binder := binding.New(nil)
	err := binder.BindAndValidate(recv, req, nil)
	assert.Nil(t, err)
	assert.Equal(t, []int{11, 12}, (**recv.X).A)
	assert.Equal(t, int32(21), (**recv.X).B)
	assert.Equal(t, &[]uint16{31, 32}, (**recv.X).C)
	assert.Equal(t, float32(41), *(**recv.X).D)
	assert.Equal(t, true, recv.Y)
	assert.Equal(t, (*int64)(nil), recv.Z)
}

func TestHeaderString(t *testing.T) {
	type Recv struct {
		X **struct {
			A []string  `header:"X-A"`
			B string    `header:"X-B"`
			C *[]string `header:"X-C,required"`
			D *string   `header:"X-D"`
		}
		Y string  `header:"X-Y,required"`
		Z *string `header:"X-Z"`
	}
	header := make(http.Header)
	header.Add("X-A", "a1")
	header.Add("X-A", "a2")
	header.Add("X-B", "b1")
	header.Add("X-C", "c1")
	header.Add("X-C", "c2")
	header.Add("X-D", "d1")
	header.Add("X-D", "d2")
	header.Add("X-Y", "y1")
	req := newRequest("", header, nil, nil)
	recv := new(Recv)
	binder := binding.New(nil)
	err := binder.BindAndValidate(recv, req, nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a1", "a2"}, (**recv.X).A)
	assert.Equal(t, "b1", (**recv.X).B)
	assert.Equal(t, []string{"c1", "c2"}, *(**recv.X).C)
	assert.Equal(t, "d1", *(**recv.X).D)
	assert.Equal(t, "y1", recv.Y)
	assert.Equal(t, (*string)(nil), recv.Z)
}

func TestHeaderNum(t *testing.T) {
	type Recv struct {
		X **struct {
			A []int     `header:"X-A"`
			B int32     `header:"X-B"`
			C *[]uint16 `header:"X-C,required"`
			D *float32  `header:"X-D"`
		}
		Y bool   `header:"X-Y,required"`
		Z *int64 `header:"X-Z"`
	}
	header := make(http.Header)
	header.Add("X-A", "11")
	header.Add("X-A", "12")
	header.Add("X-B", "21")
	header.Add("X-C", "31")
	header.Add("X-C", "32")
	header.Add("X-D", "41")
	header.Add("X-D", "42")
	header.Add("X-Y", "true")
	req := newRequest("", header, nil, nil)
	recv := new(Recv)
	binder := binding.New(nil)
	err := binder.BindAndValidate(recv, req, nil)
	assert.Nil(t, err)
	assert.Equal(t, []int{11, 12}, (**recv.X).A)
	assert.Equal(t, int32(21), (**recv.X).B)
	assert.Equal(t, &[]uint16{31, 32}, (**recv.X).C)
	assert.Equal(t, float32(41), *(**recv.X).D)
	assert.Equal(t, true, recv.Y)
	assert.Equal(t, (*int64)(nil), recv.Z)
}

func TestCookieString(t *testing.T) {
	type Recv struct {
		X **struct {
			A []string  `cookie:"a"`
			B string    `cookie:"b"`
			C *[]string `cookie:"c,required"`
			D *string   `cookie:"d"`
		}
		Y string  `cookie:"y,required"`
		Z *string `cookie:"z"`
	}
	cookies := []*http.Cookie{
		{Name: "a", Value: "a1"},
		{Name: "a", Value: "a2"},
		{Name: "b", Value: "b1"},
		{Name: "c", Value: "c1"},
		{Name: "c", Value: "c2"},
		{Name: "d", Value: "d1"},
		{Name: "d", Value: "d2"},
		{Name: "y", Value: "y1"},
	}
	req := newRequest("", nil, cookies, nil)
	recv := new(Recv)
	binder := binding.New(nil)
	err := binder.BindAndValidate(recv, req, nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a1", "a2"}, (**recv.X).A)
	assert.Equal(t, "b1", (**recv.X).B)
	assert.Equal(t, []string{"c1", "c2"}, *(**recv.X).C)
	assert.Equal(t, "d1", *(**recv.X).D)
	assert.Equal(t, "y1", recv.Y)
	assert.Equal(t, (*string)(nil), recv.Z)
}

func TestCookieNum(t *testing.T) {
	type Recv struct {
		X **struct {
			A []int     `cookie:"a"`
			B int32     `cookie:"b"`
			C *[]uint16 `cookie:"c,required"`
			D *float32  `cookie:"d"`
		}
		Y bool   `cookie:"y,required"`
		Z *int64 `cookie:"z"`
	}
	cookies := []*http.Cookie{
		{Name: "a", Value: "11"},
		{Name: "a", Value: "12"},
		{Name: "b", Value: "21"},
		{Name: "c", Value: "31"},
		{Name: "c", Value: "32"},
		{Name: "d", Value: "41"},
		{Name: "d", Value: "42"},
		{Name: "y", Value: "t"},
	}
	req := newRequest("", nil, cookies, nil)
	recv := new(Recv)
	binder := binding.New(nil)
	err := binder.BindAndValidate(recv, req, nil)
	assert.Nil(t, err)
	assert.Equal(t, []int{11, 12}, (**recv.X).A)
	assert.Equal(t, int32(21), (**recv.X).B)
	assert.Equal(t, &[]uint16{31, 32}, (**recv.X).C)
	assert.Equal(t, float32(41), *(**recv.X).D)
	assert.Equal(t, true, recv.Y)
	assert.Equal(t, (*int64)(nil), recv.Z)
}

func TestFormString(t *testing.T) {
	type Recv struct {
		X **struct {
			A []string  `form:"a"`
			B string    `form:"b"`
			C *[]string `form:"c,required"`
			D *string   `form:"d"`
		}
		Y string  `form:"y,required"`
		Z *string `form:"z"`
	}
	values := make(url.Values)
	values.Add("a", "a1")
	values.Add("a", "a2")
	values.Add("b", "b1")
	values.Add("c", "c1")
	values.Add("c", "c2")
	values.Add("d", "d1")
	values.Add("d", "d2")
	values.Add("y", "y1")
	for _, f := range []httpbody.Files{nil, {
		"f1": []httpbody.File{
			httpbody.NewFile("txt", strings.NewReader("f11 text.")),
		},
	}} {
		contentType, bodyReader := httpbody.NewFormBody2(values, f)
		header := make(http.Header)
		header.Set("Content-Type", contentType)
		req := newRequest("", header, nil, bodyReader)
		recv := new(Recv)
		binder := binding.New(nil)
		err := binder.BindAndValidate(recv, req, nil)
		assert.Nil(t, err)
		assert.Equal(t, []string{"a1", "a2"}, (**recv.X).A)
		assert.Equal(t, "b1", (**recv.X).B)
		assert.Equal(t, []string{"c1", "c2"}, *(**recv.X).C)
		assert.Equal(t, "d1", *(**recv.X).D)
		assert.Equal(t, "y1", recv.Y)
		assert.Equal(t, (*string)(nil), recv.Z)
	}
}

func TestFormNum(t *testing.T) {
	type Recv struct {
		X **struct {
			A []int     `form:"a"`
			B int32     `form:"b"`
			C *[]uint16 `form:"c,required"`
			D *float32  `form:"d"`
		}
		Y bool   `form:"y,required"`
		Z *int64 `form:"z"`
	}
	values := make(url.Values)
	values.Add("a", "11")
	values.Add("a", "12")
	values.Add("b", "-21")
	values.Add("c", "31")
	values.Add("c", "32")
	values.Add("d", "41")
	values.Add("d", "42")
	values.Add("y", "1")
	for _, f := range []httpbody.Files{nil, {
		"f1": []httpbody.File{
			httpbody.NewFile("txt", strings.NewReader("f11 text.")),
		},
	}} {
		contentType, bodyReader := httpbody.NewFormBody2(values, f)
		header := make(http.Header)
		header.Set("Content-Type", contentType)
		req := newRequest("", header, nil, bodyReader)
		recv := new(Recv)
		binder := binding.New(nil)
		err := binder.BindAndValidate(recv, req, nil)
		assert.Nil(t, err)
		assert.Equal(t, []int{11, 12}, (**recv.X).A)
		assert.Equal(t, int32(-21), (**recv.X).B)
		assert.Equal(t, &[]uint16{31, 32}, (**recv.X).C)
		assert.Equal(t, float32(41), *(**recv.X).D)
		assert.Equal(t, true, recv.Y)
		assert.Equal(t, (*int64)(nil), recv.Z)
	}
}

func TestJSON(t *testing.T) {
	type Recv struct {
		X **struct {
			A []string  `json:"a"`
			B int32     `json:""`
			C *[]uint16 `json:",required"`
			D *float32  `json:"d"`
		}
		Y string `json:"y,required"`
		Z *int64
	}

	bodyReader := strings.NewReader(`{
		"X": {
			"a": ["a1","a2"],
			"B": 21,
			"C": [31,32],
			"d": 41
		},
		"Z": 6
	}`)

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	req := newRequest("", header, nil, bodyReader)
	recv := new(Recv)
	binder := binding.New(nil)
	err := binder.BindAndValidate(recv, req, nil)
	assert.NotNil(t, err)
	assert.Equal(t, &binding.Error{ErrType: "binding failed", FailField: "y", Msg: "missing required parameter"}, err)
	assert.Equal(t, []string{"a1", "a2"}, (**recv.X).A)
	assert.Equal(t, int32(21), (**recv.X).B)
	assert.Equal(t, &[]uint16{31, 32}, (**recv.X).C)
	assert.Equal(t, float32(41), *(**recv.X).D)
	assert.Equal(t, "", recv.Y)
	assert.Equal(t, (int64)(6), *recv.Z)
}

func BenchmarkBindJSON(b *testing.B) {
	type Recv struct {
		X **struct {
			A []string `json:"a"`
			B int32
			C *[]uint16
			D *float32 `json:"d"`
		}
		Y string `json:"y"`
	}
	binder := binding.New(nil)
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	test := func() {
		bodyReader := strings.NewReader(`{
		"X": {
			"a": ["a1","a2"],
			"B": 21,
			"C": [31,32],
			"d": 41
		},
		"y": "y1"
	}`)
		req := newRequest("", header, nil, bodyReader)
		recv := new(Recv)
		err := binder.Bind(recv, req, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
	test()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		test()
	}
}

func BenchmarkStdJSON(b *testing.B) {
	type Recv struct {
		X **struct {
			A []string `json:"a"`
			B int32
			C *[]uint16
			D *float32 `json:"d"`
		}
		Y string `json:"y"`
	}
	header := make(http.Header)
	header.Set("Content-Type", "application/json")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bodyReader := strings.NewReader(`{
			"X": {
				"a": ["a1","a2"],
				"B": 21,
				"C": [31,32],
				"d": 41
			},
			"y": "y1"
		}`)

		req := newRequest("", header, nil, bodyReader)
		recv := new(Recv)
		body, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			b.Fatal(err)
		}
		err = json.Unmarshal(body, recv)
		if err != nil {
			b.Fatal(err)
		}
	}
}

type testPathParams struct{}

func (testPathParams) Get(name string) (string, bool) {
	switch name {
	case "a":
		return "a1", true
	case "b":
		return "-21", true
	case "c":
		return "31", true
	case "d":
		return "41", true
	case "y":
		return "y1", true
	case "name":
		return "henrylee2cn", true
	default:
		return "", false
	}
}

func TestPath(t *testing.T) {
	type Recv struct {
		X **struct {
			A []string  `path:"a"`
			B int32     `path:"b"`
			C *[]uint16 `path:"c,required"`
			D *float32  `path:"d"`
		}
		Y string `path:"y,required"`
		Z *int64
	}

	req := newRequest("", nil, nil, nil)
	recv := new(Recv)
	binder := binding.New(nil)
	err := binder.BindAndValidate(recv, req, new(testPathParams))
	assert.Nil(t, err)
	assert.Equal(t, []string{"a1"}, (**recv.X).A)
	assert.Equal(t, int32(-21), (**recv.X).B)
	assert.Equal(t, &[]uint16{31}, (**recv.X).C)
	assert.Equal(t, float32(41), *(**recv.X).D)
	assert.Equal(t, "y1", recv.Y)
	assert.Equal(t, (*int64)(nil), recv.Z)
}

func TestAuto(t *testing.T) {
	type Recv struct {
		X **struct {
			A []string
			B int32
			C *[]uint16 `vd:"$!=nil"`
			D *float32
		}
		Y string `vd:"$!=''"`
		Z *int64
	}
	query := make(url.Values)
	query.Add("A", "a1")
	query.Add("A", "a2")
	query.Add("B", "21")
	query.Add("Y", "y1")

	form := make(url.Values)
	form.Add("C", "31")
	form.Add("C", "32")
	form.Add("D", "41")
	form.Add("D", "42")
	for _, f := range []httpbody.Files{nil, {
		"f1": []httpbody.File{
			httpbody.NewFile("txt", strings.NewReader("f11 text.")),
		},
	}} {
		contentType, bodyReader := httpbody.NewFormBody2(form, f)
		header := make(http.Header)
		header.Set("Content-Type", contentType)
		req := newRequest("http://localhost/?"+query.Encode(), header, nil, bodyReader)
		recv := new(Recv)
		binder := binding.New(nil)
		err := binder.BindAndValidate(recv, req, nil)
		assert.Nil(t, err)
		assert.Equal(t, []string{"a1", "a2"}, (**recv.X).A)
		assert.Equal(t, int32(21), (**recv.X).B)
		assert.Equal(t, &[]uint16{31, 32}, (**recv.X).C)
		assert.Equal(t, float32(41), *(**recv.X).D)
		assert.Equal(t, "y1", recv.Y)
		assert.Equal(t, (*int64)(nil), recv.Z)
	}
}

func newRequest(u string, header http.Header, cookies []*http.Cookie, bodyReader io.Reader) *http.Request {
	if header == nil {
		header = make(http.Header)
	}
	var method = "GET"
	var body io.ReadCloser
	if bodyReader != nil {
		method = "POST"
		body = ioutil.NopCloser(bodyReader)
	}
	if u == "" {
		u = "http://localhost"
	}
	urlObj, _ := url.Parse(u)
	req := &http.Request{
		Method: method,
		URL:    urlObj,
		Body:   body,
		Header: header,
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}