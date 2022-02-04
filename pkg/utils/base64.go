package utils

import (
	"encoding/base64"
)

func DecodeB64(message string) (string, error) {
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	_, err := base64.StdEncoding.Decode(base64Text, []byte(message))
	return string(base64Text), err
}

func EcodeB64(message string) string {
	base64Text := base64.StdEncoding.Strict().EncodeToString([]byte(message))
	return base64Text
}
