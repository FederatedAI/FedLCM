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

package api

import (
	"net/http"
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/application/service"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	domainservice "github.com/FederatedAI/FedLCM/site-portal/server/domain/service"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	idKey        = "id"
	nameKey      = "name"
	uuidKey      = "uuid"
	authErrorKey = "AUTH_ERROR"
)

var authMiddleware *jwt.GinJWTMiddleware

func getKey() string {
	key := viper.GetString("siteportal.jwt.key")
	if key == "" {
		log.Warn().Msg("no pre-defined jwt key, generating a random one")
		key = rand.String(32)
	}
	return key
}

// CreateAuthMiddleware creates the authentication middleware
func CreateAuthMiddleware(repo repo.UserRepository) (err error) {
	userApp := service.UserApp{UserRepo: repo}
	authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "site portal jwt",
		Key:         []byte(getKey()),
		Timeout:     time.Hour * 24,
		MaxRefresh:  time.Hour,
		IdentityKey: idKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*service.PublicUser); ok {
				return jwt.MapClaims{
					idKey:   v.ID,
					nameKey: v.Name,
					uuidKey: v.UUID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &service.PublicUser{
				ID:   uint(claims[idKey].(float64)),
				Name: claims[nameKey].(string),
				UUID: claims[uuidKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginInfo service.LoginInfo
			if err := c.ShouldBindJSON(&loginInfo); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			if user, err := userApp.Login(&loginInfo); err == nil {
				log.Info().Msgf("user: %s logged in", loginInfo.Username)
				return user, nil
			} else if err == domainservice.ErrAccessDenied {
				return nil, err
			}
			return nil, jwt.ErrFailedAuthentication
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(code, GeneralResponse{
				Code:    0,
				Message: "",
				Data:    token,
			})
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*service.PublicUser); ok && v.Name != "" {
				if err := userApp.CheckAccess(v); err != nil {
					c.Set(authErrorKey, err)
					return false
				}
				return true
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, GeneralResponse{
				Code:    code,
				Message: message,
				Data:    nil,
			})
		},
		HTTPStatusMessageFunc: func(e error, c *gin.Context) string {
			if e == jwt.ErrForbidden {
				if v, exist := c.Get(authErrorKey); exist {
					err := v.(error)
					return err.Error()
				}
			}
			return e.Error()
		},
		TokenLookup:    "header: Authorization, cookie: jwt",
		TokenHeadName:  "Bearer",
		SendCookie:     true,
		SecureCookie:   false,
		CookieHTTPOnly: true,
		CookieName:     "jwt",
		CookieSameSite: http.SameSiteDefaultMode,
	})
	return
}
