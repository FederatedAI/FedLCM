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

package entity

import (
	"math/rand"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

var (
	alphanumericLetters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	tokenTypeMap        = map[string]RegistrationTokenType{
		"rand16": RegistrationTokenTypeRand16,
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RegistrationTokenParse pares a complete token string and returns the token type and the token string
func RegistrationTokenParse(tokenDisplayedStr string) (RegistrationTokenType, string, error) {
	splitStr := strings.SplitN(tokenDisplayedStr, ":", 2)
	if len(splitStr) == 2 {
		if t, ok := tokenTypeMap[splitStr[0]]; ok {
			return t, splitStr[1], nil
		}
	}
	return RegistrationTokenTypeUnknown, "", errors.Errorf("unknown token: %s", tokenDisplayedStr)
}

// RegistrationToken is the token entity a participant can use to register itself
type RegistrationToken struct {
	gorm.Model
	UUID        string `gorm:"type:varchar(36);index;unique"`
	Name        string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`
	TokenType   RegistrationTokenType
	TokenStr    string                           `gorm:"type:varchar(255)"`
	Repo        repo.RegistrationTokenRepository `gorm:"-"`
}

// RegistrationTokenType is an enum of token types
type RegistrationTokenType uint8

const (
	RegistrationTokenTypeUnknown RegistrationTokenType = iota
	RegistrationTokenTypeRand16
)

// RegistrationTokenOpenFL contains extra information for a token used in OpenFL federations
type RegistrationTokenOpenFL struct {
	RegistrationToken
	FederationUUID  string `gorm:"type:varchar(36)"`
	ExpiredAt       time.Time
	Limit           int
	Labels          valueobject.Labels               `gorm:"type:text"`
	ParticipantRepo repo.ParticipantOpenFLRepository `gorm:"-"`
}

// Create creates the OpenFL federation record in the repo
func (token *RegistrationTokenOpenFL) Create() error {
	if token.Name == "" {
		return errors.New("missing name")
	}
	if token.FederationUUID == "" {
		return errors.New("missing federation UUID")
	}
	if token.ExpiredAt.Before(time.Now()) {
		return errors.New("invalid expiration time")
	}
	if token.Limit <= 0 {
		return errors.New("invalid limit number")
	}
	if token.UUID == "" {
		token.UUID = uuid.NewV4().String()
	}
	if token.TokenType == RegistrationTokenTypeUnknown {
		token.TokenType = RegistrationTokenTypeRand16
	}
	if token.TokenStr == "" {
		token.TokenStr = token.TokenType.Generate()
	}
	return token.Repo.Create(token)
}

// Display returns a string representing the token and its type
func (token *RegistrationTokenOpenFL) Display() string {
	if token.TokenStr != "" {
		return token.TokenType.DisplayStr() + ":" + token.TokenStr
	}
	return ""
}

// Validate returns whether this token is still valid
func (token *RegistrationTokenOpenFL) Validate() error {
	if token.TokenStr == "" {
		return errors.New("empty token string")
	}
	if time.Now().After(token.ExpiredAt) {
		return errors.New("token expired")
	}
	if token.ParticipantRepo == nil {
		return errors.New("nil participant repo")
	}
	if count, err := token.ParticipantRepo.CountByTokenUUID(token.UUID); err != nil {
		return errors.Wrap(err, "failed to query token count")
	} else if count >= token.Limit {
		return errors.New("token limit reached")
	}
	return nil
}

// Generate generates the token string based on the token type
func (t RegistrationTokenType) Generate() string {
	switch t {
	case RegistrationTokenTypeRand16:
		b := make([]rune, 16)
		for i := range b {
			b[i] = alphanumericLetters[rand.Intn(len(alphanumericLetters))]
		}
		return string(b)
	default:
		panic("unknown RegistrationTokenType")
	}
}

// DisplayStr returns a string representing the type
func (t RegistrationTokenType) DisplayStr() string {
	switch t {
	case RegistrationTokenTypeRand16:
		return "rand16"
	default:
		return "unknown"
	}
}

func (RegistrationTokenOpenFL) TableName() string {
	// just following the gorm convention
	return "registration_token_openfls"
}
