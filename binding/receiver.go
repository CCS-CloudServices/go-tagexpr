package binding

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/bytedance/go-tagexpr"
	"github.com/bytedance/go-tagexpr/binding/jsonparam"
	"github.com/gogo/protobuf/proto"
	"github.com/henrylee2cn/goutil"
	"github.com/tidwall/gjson"
)

type in uint8

const (
	auto in = iota
	path
	query
	header
	cookie
	rawBody
	form
	json
	protobuf
	maxIn
)

var allIn = []in{
	auto,
	path,
	query,
	header,
	cookie,
	rawBody,
	form,
	json,
	protobuf,
}

type codec in

const (
	bodyUnsupport = codec(0)
	bodyForm      = codec(form)
	bodyJSON      = codec(json)
	bodyProtobuf  = codec(protobuf)
)

type receiver struct {
	hasQuery, hasCookie, hasPath, hasBody, hasVd bool

	params []*paramInfo
}

func (r *receiver) getParam(fieldSelector string) *paramInfo {
	for _, p := range r.params {
		if p.fieldSelector == fieldSelector {
			return p
		}
	}
	return nil
}

func (r *receiver) getOrAddParam(fh *tagexpr.FieldHandler, bindErrFactory func(failField, msg string) error) *paramInfo {
	fieldSelector := fh.StringSelector()
	p := r.getParam(fieldSelector)
	if p != nil {
		return p
	}
	p = new(paramInfo)
	p.fieldSelector = fieldSelector
	p.structField = fh.StructField()
	p.bindErrFactory = bindErrFactory
	r.params = append(r.params, p)
	return p
}

func (r *receiver) getBodyCodec(req *http.Request) codec {
	ct := req.Header.Get("Content-Type")
	idx := strings.Index(ct, ";")
	if idx != -1 {
		ct = strings.TrimRight(ct[:idx], " ")
	}
	switch ct {
	case "application/json":
		return bodyJSON
	case "application/x-protobuf":
		return bodyProtobuf
	case "application/x-www-form-urlencoded", "multipart/form-data":
		return bodyForm
	default:
		return bodyUnsupport
	}
}

func (r *receiver) getBody(req *http.Request) ([]byte, string, error) {
	if r.hasBody {
		bodyBytes, err := copyBody(req)
		if err == nil {
			return bodyBytes, goutil.BytesToString(bodyBytes), nil
		}
		return bodyBytes, "", nil
	}
	return nil, "", nil
}

func (r *receiver) prebindBody(structPointer interface{}, value reflect.Value, bodyCodec codec, bodyBytes []byte) error {
	switch bodyCodec {
	case bodyJSON:
		jsonparam.Assign(gjson.Parse(goutil.BytesToString(bodyBytes)), value)
	case bodyProtobuf:
		msg, ok := structPointer.(proto.Message)
		if !ok {
			return errors.New("protobuf content type is not supported")
		}
		if err := proto.Unmarshal(bodyBytes, msg); err != nil {
			return err
		}
	}
	return nil
}

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

func (r *receiver) getPostForm(req *http.Request, bodyCodec codec) (url.Values, error) {
	if bodyCodec == bodyForm && (r.hasBody) {
		if req.PostForm == nil {
			req.ParseMultipartForm(defaultMaxMemory)
		}
		return req.Form, nil
	}
	return nil, nil
}

func (r *receiver) getQuery(req *http.Request) url.Values {
	if r.hasQuery {
		return req.URL.Query()
	}
	return nil
}

func (r *receiver) getCookies(req *http.Request) []*http.Cookie {
	if r.hasCookie {
		return req.Cookies()
	}
	return nil
}

func (r *receiver) initParams() {
	names := make(map[string][maxIn]string, len(r.params))
	for _, p := range r.params {
		if p.structField.Anonymous {
			continue
		}
		a := [maxIn]string{}
		for _, paramIn := range allIn {
			a[paramIn] = p.name(paramIn)
		}
		names[p.fieldSelector] = a
	}

	for _, p := range r.params {
		paths, _ := tagexpr.FieldSelector(p.fieldSelector).Split()
		for _, info := range p.tagInfos {
			var fs string
			for _, s := range paths {
				if fs == "" {
					fs = s
				} else {
					fs = tagexpr.JoinFieldSelector(fs, s)
				}
				name := names[fs][info.paramIn]
				if name != "" {
					info.namePath = name + "."
				}
			}
			info.namePath = info.namePath + p.name(info.paramIn)
			info.requiredError = p.bindErrFactory(info.namePath, "missing required parameter")
			info.typeError = p.bindErrFactory(info.namePath, "parameter type does not match binding data")
			info.cannotError = p.bindErrFactory(info.namePath, "parameter cannot be bound")
			info.contentTypeError = p.bindErrFactory(info.namePath, "does not support binding to the content type body")
		}
	}
}