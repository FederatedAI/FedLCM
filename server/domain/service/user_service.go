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

package service

import (
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// UserService provides common services to work with entity.User
type UserService struct {
	Repo repo.UserRepository
}

// LoginService validates the provided username and password and returns the user entity when succeeded
func (s *UserService) LoginService(username, password string) (*entity.User, error) {
	user := &entity.User{Name: username}
	if err := func() error {
		if err := s.Repo.LoadByName(user); err != nil {
			return err
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		log.Warn().Err(err).Msgf("failed to validate password for user: %s", username)
		return nil, err
	}
	return user, nil
}
