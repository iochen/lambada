package utils

import (
	"encoding/base64"
	"errors"
)

func B64Decode(str string) ([]byte, error) {
	de, err := base64.StdEncoding.DecodeString(str)
	if err == nil {
		return de, err
	}

	de, err = base64.RawStdEncoding.DecodeString(str)
	if err == nil {
		return de, err
	}

	de, err = base64.URLEncoding.DecodeString(str)
	if err == nil {
		return de, err
	}

	de, err = base64.RawURLEncoding.DecodeString(str)
	if err == nil {
		return de, err
	}

	return nil, errors.New("no proper base64 decode method for: " + str)
}

func B64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
