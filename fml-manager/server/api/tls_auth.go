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

func certAuthenticator() gin.HandlerFunc {
	return func(c *gin.Context) {

		clientCert := c.Request.TLS.PeerCertificates[0]
		clientCommonName := clientCert.Subject.CommonName
		log.Info().Msgf("Request URL: %s", c.Request.URL.String())
		log.Info().Msgf("Client common name is: %s", clientCommonName)

		// TODO: validate the clientCommonName's domain is same to the fmlManagerCommonName's domain

		c.Next()
	}
}
