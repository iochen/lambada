package lambada

import (
	"net/http"
	"strings"
)

const (
	MethodGet uint8 = iota
	MethodHead
	MethodPost
	MethodPut
	MethodPatch
	MethodDelete
	MethodConnect
	MethodOptions
	MethodTrace
	MethodUnknown
)

func EncodeMethod(method string) uint8 {
	method = strings.ToUpper(method)
	switch method {
	case http.MethodGet:
		return MethodGet
	case http.MethodHead:
		return MethodHead
	case http.MethodPost:
		return MethodPost
	case http.MethodPut:
		return MethodPut
	case http.MethodPatch:
		return MethodPatch
	case http.MethodDelete:
		return MethodDelete
	case http.MethodConnect:
		return MethodConnect
	case http.MethodOptions:
		return MethodOptions
	case http.MethodTrace:
		return MethodTrace
	default:
		return MethodUnknown
	}
}

func DecodeMethod(method uint8) string {
	switch method {
	case MethodGet:
		return http.MethodGet
	case MethodHead:
		return http.MethodHead
	case MethodPost:
		return http.MethodPost
	case MethodPut:
		return http.MethodPut
	case MethodPatch:
		return http.MethodPatch
	case MethodDelete:
		return http.MethodDelete
	case MethodConnect:
		return http.MethodConnect
	case MethodOptions:
		return http.MethodOptions
	case MethodTrace:
		return http.MethodTrace
	default:
		return "UNKNOWN"
	}
}
