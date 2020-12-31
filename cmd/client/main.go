package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"

	"gopkg.in/yaml.v2"

	"github.com/iochen/lambada"
	"github.com/iochen/lambada/utils"
)

type Config struct {
	Listen      string `yaml:"listen"`
	Server      string `yaml:"server"`
	User        string `yaml:"user"`
	Key         string `yaml:"key"`
	CertFile    string `yaml:"cert"`
	CertKeyFile string `yaml:"cert_key"`
}

var cfgFile *string

func main() {
	cfgFile = flag.String("conf", "config.yaml", "config file")
	flag.Parse()

	if cfgFile == nil {
		fmt.Println("please specify a config file")
		return
	}

	confF, err := ioutil.ReadFile(*cfgFile)
	if err != nil {
		fmt.Println("An error occurred while reading config file:", err)
		return
	}

	conf := &Config{}

	err = yaml.Unmarshal(confF, &conf)
	if err != nil {
		fmt.Println("An error occurred while parse config file:", err)
		return
	}

	cert, err := ioutil.ReadFile(conf.CertFile)
	if err != nil {
		fmt.Println("An error occurred while reading cert file:", err)
		return
	}

	certKey, err := ioutil.ReadFile(conf.CertKeyFile)
	if err != nil {
		fmt.Println("An error occurred while reading cert key file:", err)
		return
	}

	goproxy.GoproxyCa, err = tls.X509KeyPair(cert, certKey)
	if err != nil {
		fmt.Println("An error occurred while parse cert:", err)
		return
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(httpReq *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		req, err := lambada.NewRequest(httpReq)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}

		buf, err := req.Encode()
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}

		lambdaReq, err := http.NewRequest("POST", conf.Server, buf)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}
		lambdaReq.Header.Add("X-User", conf.User)

		lambdaResp, err := http.DefaultClient.Do(lambdaReq)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}
		defer lambdaResp.Body.Close()

		all, err := ioutil.ReadAll(lambdaResp.Body)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}

		decode, err := utils.B64Decode(string(all))
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}

		var unzippedDecode io.Reader
		unzippedDecode, err = gzip.NewReader(bytes.NewReader(decode))
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}

		ra, err := ioutil.ReadAll(unzippedDecode)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}

		decrypt, err := utils.Decrypt([]byte(conf.Key), ra)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}

		resp, err := lambada.NewResponseFromReader(bytes.NewReader(decrypt))
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}

		httpResp, err := resp.HttpResponse()
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "error")
		}

		return nil, httpResp
	})

	log.Fatal(http.ListenAndServe(conf.Listen, proxy))
}
