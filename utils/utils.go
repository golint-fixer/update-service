/*
Copyright 2015 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// IsDirExist checks if a dir exists
func IsDirExist(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	}

	return fi.IsDir()
}

// IsFileExist checks if a file exists
func IsFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

// Contain checks if an object is included in the target object
func Contain(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)

	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	}

	return false, errors.New("not in array")
}

// EncodeBasicAuth encodes plain user/password into a base64 string
func EncodeBasicAuth(username string, password string) string {
	auth := username + ":" + password
	msg := []byte(auth)
	authorization := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
	base64.StdEncoding.Encode(authorization, msg)
	return string(authorization)
}

// DecodeBasicAuth decodes a base64 string into a plain user/password
func DecodeBasicAuth(authorization string) (username string, password string, err error) {
	basic := strings.Split(strings.TrimSpace(authorization), " ")
	if len(basic) <= 1 {
		return "", "", err
	}

	decLen := base64.StdEncoding.DecodedLen(len(basic[1]))
	decoded := make([]byte, decLen)
	authByte := []byte(basic[1])
	n, err := base64.StdEncoding.Decode(decoded, authByte)

	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", fmt.Errorf("Something went wrong decoding auth config")
	}

	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", fmt.Errorf("Invalid auth configuration file")
	}

	username = arr[0]
	password = strings.Trim(arr[1], "\x00")

	return username, password, nil
}

// ValidatePassword checks if a password is correct
func ValidatePassword(password string) error {
	if valida, _ := regexp.MatchString("[:alpha:]", password); valida != true {
		return fmt.Errorf("No alpha character in the password.")
	}

	if valida, _ := regexp.MatchString("[:digit:]", password); valida != true {
		return fmt.Errorf("No digital character in the password.")
	}

	if len(password) < 5 || len(password) > 30 {
		return fmt.Errorf("Password characters length should be between 5 - 30.")
	}

	return nil
}

// MD5 creates md5 string for an input key
func MD5(key string) string {
	md5String := fmt.Sprintf("dockyard %s is a container %d hub", key, time.Now().Unix())
	h := md5.New()
	h.Write([]byte(md5String))

	return hex.EncodeToString(h.Sum(nil))
}

// SHA512 creates sha512 string for an input data
func SHA512(body []byte) (string, error) {
	sha512h := sha512.New()
	_, err := io.Copy(sha512h, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha512h.Sum(nil)), nil
}

// Compare returns 0 if a, b are equal, -1 if a < b, other wise returns 1
func Compare(a, b string) int {
	if a == b {
		return 0
	}
	if a < b {
		return -1
	}
	return +1
}

// GenerateRSAKeyPair generate a private key and a public key
func GenerateRSAKeyPair(bits int) ([]byte, []byte, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)

	if err != nil {
		return nil, nil, err
	}

	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}
	pubBlock := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes}

	return pem.EncodeToMemory(privBlock), pem.EncodeToMemory(pubBlock), nil
}

// RSAEncrypt encrypts a content by a public key
func RSAEncrypt(keyBytes []byte, contentBytes []byte) ([]byte, error) {
	pubKey, err := getPubKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptPKCS1v15(rand.Reader, pubKey, contentBytes)
}

// RSADecrypt decrypts content by a private key
func RSADecrypt(keyBytes []byte, contentBytes []byte) ([]byte, error) {
	privKey, err := getPrivKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, privKey, contentBytes)
}

// SHA256Sign signs a content by a private key
func SHA256Sign(keyBytes []byte, contentBytes []byte) ([]byte, error) {
	privKey, err := getPrivKey(keyBytes)
	if err != nil {
		return nil, err
	}

	hashed := sha256.Sum256(contentBytes)
	return rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
}

// SHA256Verify verifies if a content is valid by a signed data an a public key
func SHA256Verify(keyBytes []byte, contentBytes []byte, signBytes []byte) error {
	pubKey, err := getPubKey(keyBytes)
	if err != nil {
		return err
	}

	signStr := hex.EncodeToString(signBytes)
	newSignBytes, _ := hex.DecodeString(signStr)
	hashed := sha256.Sum256(contentBytes)
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], newSignBytes)
}

func getPrivKey(privBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, errors.New("Fail to decode private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func getPubKey(pubBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubBytes)
	if block == nil {
		return nil, errors.New("Fail to decode public key")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("Fail get public key from public interface")
	}

	return pubKey, nil
}
