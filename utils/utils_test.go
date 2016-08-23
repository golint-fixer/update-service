/*
Copyright 2016 The ContainerOps Authors All rights reserved.

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
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testDataDir = ""
)

func init() {
	_, path, _, _ := runtime.Caller(0)
	testDataDir = filepath.Join(filepath.Dir(path), "testdata")
}

func TestIsFileExist(t *testing.T) {
	cases := []struct {
		file     string
		expected bool
	}{
		{filepath.Join(testDataDir, "hello.txt"), true},
		{filepath.Join(testDataDir, "nonexist"), false},
		{testDataDir, true},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, IsFileExist(c.file), "Fail to check file existence")
	}
}

func TestIsDirExist(t *testing.T) {
	cases := []struct {
		testDataDir string
		expected    bool
	}{
		{filepath.Join(testDataDir, "hello.txt"), false},
		{filepath.Join(testDataDir, "nonexist"), false},
		{testDataDir, true},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, IsDirExist(c.testDataDir), "Fail to check testDataDir existence")
	}
}

// TestRSAGenerateEnDe
func TestRSAGenerateEnDe(t *testing.T) {
	privBytes, pubBytes, err := GenerateRSAKeyPair(1024)
	assert.Nil(t, err, "Fail to genereate RSA Key Pair")

	testData := []byte("This is the testdata for encrypt and decryp")
	encrypted, err := RSAEncrypt(pubBytes, testData)
	assert.Nil(t, err, "Fail to encrypt data")
	decrypted, err := RSADecrypt(privBytes, encrypted)
	assert.Nil(t, err, "Fail to decrypt data")
	assert.Equal(t, testData, decrypted, "Fail to get correct data after en/de")
}

// TestSHA256Sign
func TestSHA256Sign(t *testing.T) {
	testPrivFile := filepath.Join(testDataDir, "rsa_private_key.pem")
	testContentFile := filepath.Join(testDataDir, "hello.txt")
	testSignFile := filepath.Join(testDataDir, "hello.sig")

	privBytes, _ := ioutil.ReadFile(testPrivFile)
	signBytes, _ := ioutil.ReadFile(testSignFile)
	contentBytes, _ := ioutil.ReadFile(testContentFile)
	testBytes, err := SHA256Sign(privBytes, contentBytes)
	assert.Nil(t, err, "Fail to sign")
	assert.Equal(t, testBytes, signBytes, "Fail to get valid sign data ")
}

// TestSHA256Verify
func TestSHA256Verify(t *testing.T) {
	testPubFile := filepath.Join(testDataDir, "rsa_public_key.pem")
	testContentFile := filepath.Join(testDataDir, "hello.txt")
	testSignFile := filepath.Join(testDataDir, "hello.sig")

	pubBytes, _ := ioutil.ReadFile(testPubFile)
	signBytes, _ := ioutil.ReadFile(testSignFile)
	contentBytes, _ := ioutil.ReadFile(testContentFile)
	err := SHA256Verify(pubBytes, contentBytes, signBytes)
	assert.Nil(t, err, "Fail to verify valid signed data")
	err = SHA256Verify(pubBytes, []byte("Invalid content data"), signBytes)
	assert.NotNil(t, err, "Fail to verify invalid signed data")
}

// TestSHA512
func TestSHA512(t *testing.T) {
	expectedSHA512File := filepath.Join(testDataDir, "hello.sha512")
	expectedBytes, _ := ioutil.ReadFile(expectedSHA512File)
	expected := strings.TrimSpace(string(expectedBytes))

	testContentFile := filepath.Join(testDataDir, "hello.txt")
	contentBytes, _ := ioutil.ReadFile(testContentFile)
	sha512, _ := SHA512(contentBytes)

	assert.Equal(t, expected, sha512, "Fail to create correct sha512 value")
}
