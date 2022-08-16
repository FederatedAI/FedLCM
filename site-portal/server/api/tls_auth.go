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
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// certAuthenticator is used to further check the caller's certificate
func certAuthenticator() gin.HandlerFunc {
	return func(c *gin.Context) {

		clientCommonName := c.GetHeader("X-SP-CLIENT-SDN")
		clientVerify := c.GetHeader("X-SP-CLIENT-VERIFY")
		// clientCert := c.GetHeader("X-SP-CLIENT-CERT")

		log.Info().Msgf("Request URL: %s", c.Request.URL.String())
		log.Info().Msgf("Client common name in X-SP-CLIENT-SDN is: %s", clientCommonName)
		log.Info().Msgf("Client verify result in X-SP-CLIENT-VERIFY is: %s", clientVerify)

		// TODO: validate the clientCommonName, like if its domain is same to the sitePortalCommonName's domain

		c.Next()
	}
}
