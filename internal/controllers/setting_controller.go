package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type SettingController struct {
	settingService interfaces.SettingService
}

func NewSettingController(settingService interfaces.SettingService) *SettingController {
	return &SettingController{settingService: settingService}
}

// CreateSetting godoc
// @Summary Create a new setting
// @Description Admin only. Creates a globally configurable setting.
// @Tags admin-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateSettingRequest true "Setting"
// @Success 201 {object} responses.SettingResponse "Setting created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 409 {object} responses.ErrorResponse "Setting already exists"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/settings [post]
func (c *SettingController) Create(ctx *gin.Context) {
	var req requests.CreateSettingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID := ctx.MustGet("user_id").(uuid.UUID)
	res, err := c.settingService.Create(ctx.Request.Context(), actorID, &req)
	if err != nil {
		logger.WithRequest(ctx).Error("CreateSetting failed: ", err)
		response.Conflict(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Setting created successfully", res)
}

// ListSettings godoc
// @Summary List settings
// @Description Admin only. Lists all settings, optionally filtered by category.
// @Tags admin-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category query string false "Filter by category"
// @Success 200 {array} responses.SettingResponse "Settings retrieved"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/settings [get]
func (c *SettingController) List(ctx *gin.Context) {
	category := ctx.Query("category")
	list, err := c.settingService.List(ctx.Request.Context(), category)
	if err != nil {
		logger.WithRequest(ctx).Error("ListSettings failed: ", err)
		response.InternalServerError(ctx, "Failed to list settings")
		return
	}
	response.OK(ctx, "Settings retrieved successfully", list)
}

// GetSetting godoc
// @Summary Get a setting
// @Description Admin only. Fetch a setting by key.
// @Tags admin-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "Setting key"
// @Success 200 {object} responses.SettingResponse "Setting retrieved"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/settings/{key} [get]
func (c *SettingController) Get(ctx *gin.Context) {
	key := ctx.Param("key")
	res, err := c.settingService.GetByKey(ctx.Request.Context(), key)
	if err != nil {
		response.NotFound(ctx, "Setting not found")
		return
	}
	response.OK(ctx, "Setting retrieved successfully", res)
}

// UpdateSetting godoc
// @Summary Update a setting
// @Description Admin only.
// @Tags admin-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.UpdateSettingRequest true "Update"
// @Success 200 {object} responses.SettingResponse "Setting updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/settings [put]
func (c *SettingController) Update(ctx *gin.Context) {
	var req requests.UpdateSettingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID := ctx.MustGet("user_id").(uuid.UUID)
	res, err := c.settingService.Update(ctx.Request.Context(), actorID, &req)
	if err != nil {
		if err.Error() == "setting not found" {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalServerError(ctx, "Failed to update setting")
		return
	}
	response.OK(ctx, "Setting updated successfully", res)
}

// BulkUpdateSettings godoc
// @Summary Bulk update settings
// @Description Admin only. Update multiple settings in one request.
// @Tags admin-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.BulkUpdateSettingsRequest true "Bulk update"
// @Success 200 {array} responses.SettingResponse "Settings updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/settings/bulk [put]
func (c *SettingController) BulkUpdate(ctx *gin.Context) {
	var req requests.BulkUpdateSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID := ctx.MustGet("user_id").(uuid.UUID)
	res, err := c.settingService.BulkUpdate(ctx.Request.Context(), actorID, &req)
	if err != nil {
		logger.WithRequest(ctx).Error("BulkUpdateSettings failed: ", err)
		response.InternalServerError(ctx, "Failed to bulk update settings")
		return
	}
	response.OK(ctx, "Settings updated successfully", res)
}

// DeleteSetting godoc
// @Summary Delete a setting
// @Description Admin only.
// @Tags admin-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "Setting key"
// @Success 200 {object} responses.Response "Setting deleted"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/settings/{key} [delete]
func (c *SettingController) Delete(ctx *gin.Context) {
	key := ctx.Param("key")
	if err := c.settingService.Delete(ctx.Request.Context(), key); err != nil {
		response.InternalServerError(ctx, "Failed to delete setting")
		return
	}
	response.OK(ctx, "Setting deleted successfully", nil)
}

// PublicSettings godoc
// @Summary Get public settings
// @Description Public endpoint. Returns settings marked as public.
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} responses.PublicSettingsResponse "Public settings"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /settings/public [get]
func (c *SettingController) PublicSettings(ctx *gin.Context) {
	res, err := c.settingService.ListPublic(ctx.Request.Context())
	if err != nil {
		response.InternalServerError(ctx, "Failed to load public settings")
		return
	}
	response.OK(ctx, "Public settings retrieved successfully", res)
}
