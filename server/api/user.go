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
	"strconv"

	"github.com/FederatedAI/FedLCM/server/application/service"
	"github.com/FederatedAI/FedLCM/server/constants"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// UserController manages user related API calls
type UserController struct {
	userAppService *service.UserApp
}

// NewUserController returns a controller instance to handle user API requests
func NewUserController(repo repo.UserRepository) *UserController {
	return &UserController{
		userAppService: &service.UserApp{
			UserRepo: repo,
		},
	}
}

// Route set up route mappings to user related APIs
func (controller *UserController) Route(r *gin.RouterGroup) {
	users := r.Group("user")
	{
		users.POST("/login", controller.login)
		users.POST("/logout", controller.logout)
	}
	users.Use(authMiddleware.MiddlewareFunc())
	{
		users.GET("/current", controller.getCurrentUsername)
		users.PUT("/:id/password", controller.updatePassword)
	}
}

// login to lifecycle manager using the provided credentials
// @Summary login to lifecycle manager
// @Tags    User
// @Produce json
// @Param   credentials body     service.LoginInfo true "credentials for login"
// @Success 200         {object} GeneralResponse   "Success"
// @Failure 401         {object} GeneralResponse   "Unauthorized operation"
// @Router  /user/login [post]
func (controller *UserController) login(c *gin.Context) {
	authMiddleware.LoginHandler(c)
}

// logout from the lifecycle manager
// @Summary logout from the lifecycle manager
// @Tags    User
// @Produce json
// @Success 200 {object} GeneralResponse           "Success"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router  /user/logout [post]
func (controller *UserController) logout(c *gin.Context) {
	authMiddleware.LogoutHandler(c)
}

// getCurrentUser return current user
// @Summary Return current user in the jwt token
// @Tags    User
// @Produce json
// @Success 200 {object} GeneralResponse{data=string} "Success, the name of current user"
// @Failure 401 {object} GeneralResponse              "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int}    "Internal server error"
// @Router  /user/current [get]
func (controller *UserController) getCurrentUsername(c *gin.Context) {
	if username, err := func() (string, error) {
		claims := jwt.ExtractClaims(c)
		// the auth middleware makes sure username exists
		username := claims[nameKey].(string)
		return username, nil
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: username,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// updatePassword update user password
// @Summary Update user password
// @Tags    User
// @Produce json
// @Param   passwordChangeInfo body     service.PwdChangeInfo     string "current and new password"
// @Success 200                {object} GeneralResponse           "Success"
// @Failure 401                {object} GeneralResponse           "Unauthorized operation"
// @Failure 500                {object} GeneralResponse{code=int} "Internal server error"
// @Router  /user/{userId}/password [put]
func (controller *UserController) updatePassword(c *gin.Context) {
	if err := func() error {
		userId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}
		passwordChangeInfo := &service.PwdChangeInfo{}
		if err := c.ShouldBindJSON(&passwordChangeInfo); err != nil {
			return err
		}
		return controller.userAppService.UpdateUserPassword(userId, passwordChangeInfo)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
		}
		c.JSON(http.StatusOK, resp)
	}
}
