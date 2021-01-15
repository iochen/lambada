package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/iochen/lambada"
	"github.com/iochen/lambada/utils"
)

type Event struct {
	Body              string              `json:"body"`
	Path              string              `json:"path"`
	HTTPMethod        string              `json:"httpMethod"`
	IsBase64Encoded   bool                `json:"isBase64Encoded"`
	Headers           map[string]string   `json:"headers"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`
}

func main() {
	lambda.Start(Handler)
}

func Handler(ctx context.Context, event Event) (string, error) {
	if event.Headers["x-user"] != os.Getenv("LBD_USER") {
		return "", errors.New("not valid session")
	}

	var body []byte
	var err error
	if event.IsBase64Encoded {
		body, err = utils.B64Decode(event.Body)
		if err != nil {
			return "", err
		}
	} else {
		body = []byte(event.Body)
	}

	unzippedReader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	unzippedReq, err := ioutil.ReadAll(unzippedReader)
	if err != nil {
		return "", err
	}

	decryptedReq, err := utils.Decrypt([]byte(os.Getenv("LBD_KEY")), unzippedReq)
	if err != nil {
		return "", err
	}

	request, err := lambada.NewRequestFromReader(bytes.NewReader(decryptedReq))
	if err != nil {
		return "", err
	}

	req, err := request.HttpRequest()
	if err != nil {
		return "", err
	}
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()

	resp, err := lambada.NewResponse(httpResp)
	if err != nil {
		return "", err
	}

	buf, err := resp.Encode()
	if err != nil {
		return "", err
	}

	encrypted, err := utils.Encrypt([]byte(os.Getenv("LBD_KEY")), buf.Bytes())
	if err != nil {
		return "", err
	}

	zippedBuf := &bytes.Buffer{}
	w := gzip.NewWriter(zippedBuf)
	_, err = w.Write(encrypted)
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}

	en := utils.B64Encode(zippedBuf.Bytes())

	return en, nil
}
