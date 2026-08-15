package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"
)

type UserController struct {
	userService interfaces.UserService
}

func NewUserController(userService interfaces.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.UserResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/profile [get]
func (c *UserController) GetProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Unauthorized(ctx, "User not authenticated", nil)
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		response.InternalServerError(ctx, "Invalid user ID format")
		return
	}

	result, err := c.userService.GetProfile(ctx.Request.Context(), uid)
	if err != nil {
		logger.WithRequest(ctx).Error("GetProfile failed: ", err)
		if err.Error() == "user not found" {
			response.NotFound(ctx, "User not found")
			return
		}
		response.InternalServerError(ctx, "Failed to get profile")
		return
	}

	response.OK(ctx, "Profile retrieved successfully", result)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} responses.UserResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/profile [put]
func (c *UserController) UpdateProfile(ctx *gin.Context) {
	var req requests.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Unauthorized(ctx, "User not authenticated", nil)
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		response.InternalServerError(ctx, "Invalid user ID format")
		return
	}

	result, err := c.userService.UpdateProfile(ctx.Request.Context(), uid, &req)
	if err != nil {
		logger.WithRequest(ctx).Error("UpdateProfile failed: ", err)
		response.InternalServerError(ctx, "Failed to update profile")
		return
	}

	response.OK(ctx, "Profile updated successfully", result)
}

// UpdateRole godoc
// @Summary Update user role
// @Description Add a new role to user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.UpdateRoleRequest true "Update role request"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/role [post]
func (c *UserController) UpdateRole(ctx *gin.Context) {
	var req requests.UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Unauthorized(ctx, "User not authenticated", nil)
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		response.InternalServerError(ctx, "Invalid user ID format")
		return
	}

	if err := c.userService.UpdateRole(ctx.Request.Context(), uid, &req); err != nil {
		logger.WithRequest(ctx).Error("UpdateRole failed: ", err)
		if err.Error() == "role not found" {
			response.NotFound(ctx, "Role not found")
			return
		}
		response.InternalServerError(ctx, "Failed to update role")
		return
	}

	response.OK(ctx, "Role updated successfully", nil)
}

// SubmitKYC godoc
// @Summary Submit KYC verification
// @Description Submit KYC documents for verification
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.KYCRequest true "KYC request"
// @Success 201 {object} responses.KYCResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/kyc [post]
func (c *UserController) SubmitKYC(ctx *gin.Context) {
	var req requests.KYCRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Unauthorized(ctx, "User not authenticated", nil)
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		response.InternalServerError(ctx, "Invalid user ID format")
		return
	}

	result, err := c.userService.SubmitKYC(ctx.Request.Context(), uid, &req)
	if err != nil {
		logger.WithRequest(ctx).Error("SubmitKYC failed: ", err)
		response.InternalServerError(ctx, "Failed to submit KYC")
		return
	}

	response.Created(ctx, "KYC submitted successfully", result)
}

// GetKYCStatus godoc
// @Summary Get KYC status
// @Description Get KYC verification status
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.KYCResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/kyc/status [get]
func (c *UserController) GetKYCStatus(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Unauthorized(ctx, "User not authenticated", nil)
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		response.InternalServerError(ctx, "Invalid user ID format")
		return
	}

	result, err := c.userService.GetKYCStatus(ctx.Request.Context(), uid)
	if err != nil {
		logger.WithRequest(ctx).Error("GetKYCStatus failed: ", err)
		response.InternalServerError(ctx, "Failed to get KYC status")
		return
	}

	response.OK(ctx, "KYC status retrieved successfully", result)
}
