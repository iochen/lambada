package lambada

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"net/http"
)

type Request struct {
	URLLength    uint16
	URL          []byte
	Method       uint8
	HeaderLength uint16
	Header       []byte
	Body         []byte
}

func NewRequest(httpReq *http.Request) (*Request, error) {
	var body []byte
	if httpReq.Body != nil {
		var err error
		body, err = ioutil.ReadAll(httpReq.Body)
		if err != nil {
			return &Request{}, err
		}
	}

	url := httpReq.URL.String()
	urlL := len(url)
	if urlL > math.MaxUint16 {
		return &Request{}, errors.New("url too long")
	}

	header, err := json.Marshal(httpReq.Header)
	if err != nil {
		return &Request{}, err
	}
	headerL := len(header)
	if headerL > math.MaxUint16 {
		return &Request{}, errors.New("header too long")
	}

	return &Request{
		URLLength:    uint16(urlL),
		URL:          []byte(url),
		Method:       EncodeMethod(httpReq.Method),
		HeaderLength: uint16(headerL),
		Header:       []byte(header),
		Body:         body,
	}, nil
}

func NewRequestFromReader(r io.Reader) (*Request, error) {
	req := &Request{}
	err := binary.Read(r, binary.LittleEndian, &req.URLLength)
	if err != nil {
		return &Request{}, err
	}
	url := make([]byte, req.URLLength)
	_, err = io.ReadFull(r, url)
	if err != nil {
		return &Request{}, err
	}
	req.URL = url
	err = binary.Read(r, binary.LittleEndian, &req.Method)
	if err != nil {
		return &Request{}, err
	}
	err = binary.Read(r, binary.LittleEndian, &req.HeaderLength)
	if err != nil {
		return &Request{}, err
	}
	header := make([]byte, req.HeaderLength)
	_, err = io.ReadFull(r, header)
	if err != nil {
		return &Request{}, err
	}
	req.Header = header
	req.Body, err = ioutil.ReadAll(r)
	if err != nil {
		return &Request{}, err
	}
	return req, nil
}

func (req *Request) Encode() (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	url := req.URL
	urlL := uint16(len(url))
	if urlL > math.MaxUint16 {
		return nil, errors.New("url too long")
	}
	header := req.Header
	headerL := uint16(len(header))
	if headerL > math.MaxUint16 {
		return nil, errors.New("header too long")
	}

	// URL Length
	err := binary.Write(buf, binary.LittleEndian, urlL)
	if err != nil {
		return nil, err
	}

	// URL
	buf.Write(url)

	// Method
	buf.WriteByte(req.Method)

	// Header Length
	err = binary.Write(buf, binary.LittleEndian, headerL)
	if err != nil {
		return nil, err
	}

	// Header
	buf.Write(header)

	// Body
	buf.Write(req.Body)

	return buf, nil
}

func (req *Request) HttpRequest() (*http.Request, error) {
	method := DecodeMethod(req.Method)
	header := map[string][]string{}
	err := json.Unmarshal(req.Header, &header)
	if err != nil {
		return &http.Request{}, err
	}
	bodyR := bytes.NewBuffer(req.Body)
	httpReq, err := http.NewRequest(method, string(req.URL), bodyR)
	if err != nil {
		return &http.Request{}, err
	}
	httpReq.Header = header
	return httpReq, nil
}
