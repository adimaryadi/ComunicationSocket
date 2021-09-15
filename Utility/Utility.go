package Utility

import (
	"github.com/Luzifer/go-openssl"
	"github.com/asim/go-micro/util/log"
	"strings"
)
type DigestFunc func([]byte) []byte


func Dencrytion(key string,ciphertext string) string {
	openSSL := openssl.New()
	decription, err := openSSL.DecryptBytes(key, []byte(ciphertext), openssl.BytesToKeyMD5)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Terjemaah =>",string(decription))
	return string(decription)
}


func SpaceFieldsJoin(str string) string {
	return strings.Join(strings.Fields(str), "+")
}