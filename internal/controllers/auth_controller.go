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

type AuthController struct {
	authService interfaces.AuthService
}

func NewAuthController(authService interfaces.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// SignUp godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.SignUpRequest true "Sign up request"
// @Success 201 {object} responses.AuthResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/signup [post]
func (c *AuthController) SignUp(ctx *gin.Context) {
	var req requests.SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	// Call service
	result, err := c.authService.SignUp(ctx.Request.Context(), &req)
	if err != nil {
		logger.WithRequest(ctx).Error("SignUp failed: ", err)
		if err.Error() == "email already registered" {
			response.Conflict(ctx, "Email already registered", nil)
			return
		}
		response.InternalServerError(ctx, "Failed to create account")
		return
	}

	response.Created(ctx, "Account created successfully", result)
}

// SignIn godoc
// @Summary Sign in a user
// @Description Authenticate user and get tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.SignInRequest true "Sign in request"
// @Success 200 {object} responses.AuthResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/signin [post]
func (c *AuthController) SignIn(ctx *gin.Context) {
	var req requests.SignInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	// Call service
	result, err := c.authService.SignIn(ctx.Request.Context(), &req)
	if err != nil {
		logger.WithRequest(ctx).Warning("SignIn failed: ", err)
		switch err.Error() {
		case "invalid email or password":
			response.Unauthorized(ctx, "Invalid email or password", nil)
		case "email not verified":
			response.Unauthorized(ctx, "Email not verified", nil)
		case "account is not active":
			response.Unauthorized(ctx, "Account is not active", nil)
		default:
			response.InternalServerError(ctx, "Failed to sign in")
		}
		return
	}

	response.OK(ctx, "Signed in successfully", result)
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify email with 6-digit code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.VerifyEmailRequest true "Verify email request"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/verify-email [post]
func (c *AuthController) VerifyEmail(ctx *gin.Context) {
	var req requests.VerifyEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	// Call service
	if err := c.authService.VerifyEmail(ctx.Request.Context(), &req); err != nil {
		logger.WithRequest(ctx).Warning("VerifyEmail failed: ", err)
		switch err.Error() {
		case "invalid verification code":
			response.BadRequest(ctx, "Invalid verification code", nil)
		case "verification code expired":
			response.BadRequest(ctx, "Verification code expired", nil)
		case "too many attempts, please resend code":
			response.BadRequest(ctx, "Too many attempts, please resend code", nil)
		case "email already verified":
			response.Conflict(ctx, "Email already verified", nil)
		default:
			response.InternalServerError(ctx, "Failed to verify email")
		}
		return
	}

	response.OK(ctx, "Email verified successfully", nil)
}

// ResendVerification godoc
// @Summary Resend verification code
// @Description Resend email verification code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.ResendVerificationRequest true "Resend verification request"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/resend-verification [post]
func (c *AuthController) ResendVerification(ctx *gin.Context) {
	var req requests.ResendVerificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	// Call service
	if err := c.authService.ResendVerification(ctx.Request.Context(), &req); err != nil {
		logger.WithRequest(ctx).Warning("ResendVerification failed: ", err)
		if err.Error() == "email already verified" {
			response.Conflict(ctx, "Email already verified", nil)
			return
		}
		response.InternalServerError(ctx, "Failed to resend verification code")
		return
	}

	response.OK(ctx, "Verification code sent successfully", nil)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} responses.TokenResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/refresh-token [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req requests.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	// Call service
	result, err := c.authService.RefreshToken(ctx.Request.Context(), &req)
	if err != nil {
		logger.WithRequest(ctx).Warning("RefreshToken failed: ", err)
		response.Unauthorized(ctx, "Invalid refresh token", nil)
		return
	}

	response.OK(ctx, "Token refreshed successfully", result)
}

// Logout godoc
// @Summary Logout user
// @Description Revoke refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.LogoutRequest true "Logout request"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	var req requests.LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	// Call service
	if err := c.authService.Logout(ctx.Request.Context(), &req); err != nil {
		logger.WithRequest(ctx).Warning("Logout failed: ", err)
		response.InternalServerError(ctx, "Failed to logout")
		return
	}

	response.OK(ctx, "Logged out successfully", nil)
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Send password reset email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/forgot-password [post]
func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var req requests.ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	// Call service
	if err := c.authService.ForgotPassword(ctx.Request.Context(), &req); err != nil {
		logger.WithRequest(ctx).Error("ForgotPassword failed: ", err)
		response.InternalServerError(ctx, "Failed to process request")
		return
	}

	// Always return success (don't reveal if email exists)
	response.OK(ctx, "If the email exists, a password reset link has been sent", nil)
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password using token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	var req requests.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	// Call service
	if err := c.authService.ResetPassword(ctx.Request.Context(), &req); err != nil {
		logger.WithRequest(ctx).Warning("ResetPassword failed: ", err)
		if err.Error() == "invalid or expired reset token" {
			response.BadRequest(ctx, "Invalid or expired reset token", nil)
			return
		}
		response.InternalServerError(ctx, "Failed to reset password")
		return
	}

	response.OK(ctx, "Password reset successfully", nil)
}

// ChangePassword godoc
// @Summary Change password
// @Description Change password for authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.ChangePasswordRequest true "Change password request"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 422 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/change-password [post]
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	var req requests.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Unauthorized(ctx, "User not authenticated", nil)
		return
	}

	// Convert user ID to string
	userIDStr, ok := userID.(uuid.UUID)
	if !ok {
		response.InternalServerError(ctx, "Invalid user ID format")
		return
	}

	// Call service
	if err := c.authService.ChangePassword(ctx.Request.Context(), userIDStr.String(), &req); err != nil {
		logger.WithRequest(ctx).Warning("ChangePassword failed: ", err)
		switch err.Error() {
		case "current password is incorrect":
			response.BadRequest(ctx, "Current password is incorrect", nil)
		case "new password must be different from current password":
			response.BadRequest(ctx, "New password must be different from current password", nil)
		default:
			response.InternalServerError(ctx, "Failed to change password")
		}
		return
	}

	response.OK(ctx, "Password changed successfully", nil)
}
