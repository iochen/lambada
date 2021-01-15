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
	"os"
	"path"

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

	var confPath string
	ucd, ucdErr := os.UserConfigDir()
	confF, err := ioutil.ReadFile(*cfgFile)
	confPath = path.Dir(*cfgFile)
	if err != nil && ucdErr != nil {
		fmt.Printf("An error occurred while reading config file at %s: %v.\n", *cfgFile, err)
		os.Exit(1)
		return
	} else if err != nil {
		lUcd := path.Join(ucd, "lambada/")
		confF, err = ioutil.ReadFile(path.Join(lUcd, "config.yaml"))
		if err != nil {
			fmt.Printf("An error occurred while reading config file at %s or %s: %v.\n", *cfgFile, ucd, err)
			os.Exit(1)
			return
		}
		confPath = lUcd
	}

	conf := &Config{}

	err = yaml.Unmarshal(confF, &conf)
	if err != nil {
		fmt.Println("An error occurred while parsing config file:", err)
		os.Exit(1)
		return
	}

	cert, err := ioutil.ReadFile(path.Join(confPath, conf.CertFile))
	if err != nil {
		fmt.Println("An error occurred while reading cert file:", err)
		os.Exit(1)
		return
	}

	certKey, err := ioutil.ReadFile(path.Join(confPath, conf.CertKeyFile))
	if err != nil {
		fmt.Println("An error occurred while reading cert key file:", err)
		os.Exit(1)
		return
	}

	goproxy.GoproxyCa, err = tls.X509KeyPair(cert, certKey)
	if err != nil {
		fmt.Println("An error occurred while parsing cert:", err)
		os.Exit(1)
		return
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(httpReq *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		req, err := lambada.NewRequest(httpReq)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when parsing request.")
		}

		encodedReq, err := req.Encode()
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when encoding request.")
		}

		encryptedReq, err := utils.Encrypt([]byte(conf.Key), encodedReq.Bytes())
		if err != nil {
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when encrypting request.")
		}

		zippedReq := &bytes.Buffer{}
		zipWriter := gzip.NewWriter(zippedReq)
		_, err = zipWriter.Write(encryptedReq)
		if err != nil {
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when zipping request.")
		}

		if err := zipWriter.Close(); err != nil {
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when closing gzip writer.")
		}

		lambdaReq, err := http.NewRequest("POST", conf.Server, zippedReq)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when getting http request.")
		}

		lambdaReq.Header.Add("X-User", conf.User)

		lambdaResp, err := http.DefaultClient.Do(lambdaReq)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when posting http request.")
		}
		defer lambdaResp.Body.Close()

		all, err := ioutil.ReadAll(lambdaResp.Body)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when reading http response.")
		}

		decode, err := utils.B64Decode(string(all))
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when decoding http response.")
		}

		var unzippedDecode io.Reader
		unzippedDecode, err = gzip.NewReader(bytes.NewReader(decode))
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when unzipping response.")
		}

		ra, err := ioutil.ReadAll(unzippedDecode)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when reading unzipped response.")
		}

		decrypt, err := utils.Decrypt([]byte(conf.Key), ra)
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when decrypting response.")
		}

		resp, err := lambada.NewResponseFromReader(bytes.NewReader(decrypt))
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when parsing response.")
		}

		httpResp, err := resp.HttpResponse()
		if err != nil {
			log.Println(err)
			return &http.Request{}, goproxy.TextResponse(httpReq, "Lambada: An error occurred when getting response.")
		}

		return nil, httpResp
	})

	log.Fatal(http.ListenAndServe(conf.Listen, proxy))
}
