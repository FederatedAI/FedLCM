// Copyright 2022 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/spf13/viper"
)

const (
	encryptHeaderV1 = "<enc-v1>"
)

// Encrypt encrypts the passed secret using AES with configured secret key
func Encrypt(secret string) (string, error) {
	if len(secret) == 0 {
		return secret, nil
	}
	secretKey, _ := findSecretKey()
	encrypted, err := reversibleEncrypt(secret, secretKey)
	if err != nil {
		return "", err
	}
	return encrypted, nil
}

// Decrypt decrypts the passed secret using configured secret key
func Decrypt(secret string) (string, error) {
	if len(secret) == 0 {
		return "", nil
	}
	secretKey, _ := findSecretKey()
	decrypted, err := reversibleDecrypt(secret, secretKey)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}

func findSecretKey() (string, error) {
	// TODO: check key size as AES only supports key sizes of 16, 24 or 32 bytes
	secretKey := viper.GetString("lifecyclemanager.secretkey")
	if secretKey == "" {
		secretKey = "passphrase123456"
	}
	return secretKey, nil
}

// reversibleEncrypt encrypts the str with aes/base64
func reversibleEncrypt(str, key string) (string, error) {
	keyBytes := []byte(key)
	var block cipher.Block
	var err error

	if block, err = aes.NewCipher(keyBytes); err != nil {
		return "", err
	}

	// ensures the value is no larger than 64 MB, which fits comfortably within an int and avoids the overflow
	if len(str) > 64*1024*1024 {
		return "", errors.New("str value too large")
	}

	size := aes.BlockSize + len(str)
	cipherText := make([]byte, size)
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(str))
	encrypted := encryptHeaderV1 + base64.RawURLEncoding.EncodeToString(cipherText)
	return encrypted, nil
}

// reversibleDecrypt decrypts the str with aes/base64 or base 64 depending on "header"
func reversibleDecrypt(str, key string) (string, error) {
	if strings.HasPrefix(str, encryptHeaderV1) {
		str = str[len(encryptHeaderV1):]
		return decryptAES(str, key)
	}
	// fallback to base64
	return decodeB64(str)
}

func decodeB64(str string) (string, error) {
	cipherText, err := base64.RawURLEncoding.DecodeString(str)
	return string(cipherText), err
}

func decryptAES(str, key string) (string, error) {
	keyBytes := []byte(key)
	var block cipher.Block
	var cipherText []byte
	var err error

	if block, err = aes.NewCipher(keyBytes); err != nil {
		return "", err
	}
	if cipherText, err = base64.RawURLEncoding.DecodeString(str); err != nil {
		return "", err
	}
	if len(cipherText) < aes.BlockSize {
		err = errors.New("cipherText too short")
		return "", err
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)
	return string(cipherText), nil
}
