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

	"github.com/FederatedAI/FedLCM/site-portal/server/application/service"
	"github.com/FederatedAI/FedLCM/site-portal/server/constants"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
		users.GET("", controller.listUsers)
		users.GET("/current", controller.getCurrentUsername)
		users.PUT("/:id/permission", controller.updatePermission)
		users.PUT("/:id/password", controller.updatePassword)
	}
}

// listUsers list all users
// @Summary List all saved users
// @Tags User
// @Produce json
// @Success 200 {object} GeneralResponse{data=[]service.PublicUser} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /user [get]
func (controller *UserController) listUsers(c *gin.Context) {
	users, err := controller.userAppService.GetUsers()
	if err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
			Data:    nil,
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: users,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// updatePermission update user permission
// @Summary Update user permission
// @Tags User
// @Produce json
// @Param permission body valueobject.UserPermissionInfo true "Permission, must contain all permissions, otherwise the missing once will be considered as false"
// @Param id path string true "User ID"
// @Success 200 {object} GeneralResponse "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /user/{id}/permission [put]
func (controller *UserController) updatePermission(c *gin.Context) {
	if data, err := func() (interface{}, error) {
		userId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return nil, err
		}
		user := &service.PublicUser{
			ID: uint(userId),
		}
		if err := c.ShouldBindJSON(&user.UserPermissionInfo); err != nil {
			return nil, err
		}
		claims := jwt.ExtractClaims(c)
		// XXX: enhance this simple check
		username := claims[nameKey].(string)
		if username != "Admin" {
			return nil, errors.New("only Admin user can change permissions")
		}
		return nil, controller.userAppService.UpdateUserPermission(user)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
			Data:    nil,
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: data,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// login login to site portal using the provided credentials
// @Summary login to site portal
// @Tags User
// @Produce json
// @Param credentials body service.LoginInfo true "credentials for login"
// @Success 200 {object} GeneralResponse "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Router /user/login [post]
func (controller *UserController) login(c *gin.Context) {
	authMiddleware.LoginHandler(c)
}

// logout logout from the site portal
// @Summary logout from the site portal
// @Tags User
// @Produce json
// @Success 200 {object} GeneralResponse "Success"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /user/logout [post]
func (controller *UserController) logout(c *gin.Context) {
	authMiddleware.LogoutHandler(c)
}

// getCurrentUser return current user
// @Summary Return current user in the jwt token
// @Tags User
// @Produce json
// @Success 200 {object} GeneralResponse{data=string} "Success, the name of current user"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /user/current [get]
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
// @Summary Update user Password
// @Tags User
// @Produce json
// @Param passwordChangeInfo body service.PwdChangeInfo string "current and new password"
// @Success 200 {object} GeneralResponse "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /user/{id}/password [put]
func (controller *UserController) updatePassword(c *gin.Context) {
	if err := func() error {
		userId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}
		claims := jwt.ExtractClaims(c)
		// the auth middleware makes sure username exists
		currentId := int(claims[idKey].(float64))
		if userId != currentId {
			return errors.New("invalid user id")
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
