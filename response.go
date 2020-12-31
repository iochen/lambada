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

type Response struct {
	StatusCode   int16
	StatusLength uint8
	Status       []byte
	HeaderLength uint16
	Header       []byte
	Body         []byte
}

func NewResponse(httpResp *http.Response) (*Response, error) {
	var body []byte
	if httpResp.Body != nil {
		var err error
		body, err = ioutil.ReadAll(httpResp.Body)
		if err != nil {
			return &Response{}, err
		}
	}

	statusCode := httpResp.StatusCode
	status := httpResp.Status
	statusL := len(status)
	if statusL > math.MaxUint8 {
		return &Response{}, errors.New("status too long")
	}

	header, err := json.Marshal(httpResp.Header)
	if err != nil {
		return &Response{}, err
	}
	headerL := len(header)
	if headerL > math.MaxUint16 {
		return &Response{}, errors.New("header too long")
	}

	return &Response{
		StatusCode:   int16(statusCode),
		StatusLength: uint8(statusL),
		Status:       []byte(status),
		HeaderLength: uint16(headerL),
		Header:       []byte(header),
		Body:         body,
	}, nil
}

func NewResponseFromReader(r io.Reader) (*Response, error) {
	resp := &Response{}
	err := binary.Read(r, binary.LittleEndian, &resp.StatusCode)
	if err != nil {
		return &Response{}, err
	}
	err = binary.Read(r, binary.LittleEndian, &resp.StatusLength)
	if err != nil {
		return &Response{}, err
	}
	status := make([]byte, resp.StatusLength)
	_, err = io.ReadFull(r, status)
	if err != nil {
		return &Response{}, err
	}
	resp.Status = status
	err = binary.Read(r, binary.LittleEndian, &resp.HeaderLength)
	if err != nil {
		return &Response{}, err
	}
	header := make([]byte, resp.HeaderLength)
	_, err = io.ReadFull(r, header)
	if err != nil {
		return &Response{}, err
	}
	resp.Header = header
	resp.Body, err = ioutil.ReadAll(r)
	if err != nil {
		return &Response{}, err
	}
	return resp, nil
}

func (resp *Response) Encode() (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	status := resp.Status
	statusL := len(status)
	if statusL > math.MaxUint8 {
		return nil, errors.New("header too long")
	}
	header := resp.Header
	headerL := uint16(len(header))
	if headerL > math.MaxUint16 {
		return nil, errors.New("header too long")
	}

	// status code
	err := binary.Write(buf, binary.LittleEndian, resp.StatusCode)
	if err != nil {
		return nil, err
	}

	// status length
	buf.WriteByte(uint8(statusL))

	// status
	buf.Write(status)

	// Header Length
	err = binary.Write(buf, binary.LittleEndian, headerL)
	if err != nil {
		return nil, err
	}

	// Header
	buf.Write(header)

	// Body
	buf.Write(resp.Body)

	return buf, nil
}

func (resp *Response) HttpResponse() (*http.Response, error) {
	statusCode := resp.StatusCode
	status := string(resp.Status)
	header := map[string][]string{}
	err := json.Unmarshal(resp.Header, &header)
	if err != nil {
		return &http.Response{}, err
	}
	bodyR := bytes.NewBuffer(resp.Body)
	return &http.Response{
		StatusCode: int(statusCode),
		Status:     status,
		Header:     header,
		Body:       ioutil.NopCloser(bodyR),
	}, nil
}
