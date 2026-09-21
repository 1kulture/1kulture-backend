package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type CommissionController struct {
	commissionService interfaces.CommissionService
}

func NewCommissionController(cs interfaces.CommissionService) *CommissionController {
	return &CommissionController{commissionService: cs}
}

// CreateTier godoc
// @Summary Create commission tier
// @Description Admin only.
// @Tags admin-commission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateCommissionTierRequest true "Tier"
// @Success 201 {object} responses.CommissionTierResponse "Tier created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/commission/tiers [post]
func (c *CommissionController) CreateTier(ctx *gin.Context) {
	var req requests.CreateCommissionTierRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	res, err := c.commissionService.CreateTier(ctx.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Created(ctx, "Commission tier created", res)
}

// ListTiers godoc
// @Summary List commission tiers
// @Description Admin only.
// @Tags admin-commission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param active_only query bool false "Only active"
// @Success 200 {array} responses.CommissionTierResponse "Tiers"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/commission/tiers [get]
func (c *CommissionController) ListTiers(ctx *gin.Context) {
	onlyActive, _ := strconv.ParseBool(ctx.DefaultQuery("active_only", "false"))
	list, err := c.commissionService.ListTiers(ctx.Request.Context(), onlyActive)
	if err != nil {
		response.InternalServerError(ctx, "Failed to list tiers")
		return
	}
	response.OK(ctx, "Tiers retrieved successfully", list)
}

// UpdateTier godoc
// @Summary Update commission tier
// @Description Admin only.
// @Tags admin-commission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tier ID"
// @Param request body requests.UpdateCommissionTierRequest true "Update"
// @Success 200 {object} responses.CommissionTierResponse "Tier updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/commission/tiers/{id} [put]
func (c *CommissionController) UpdateTier(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid tier id", nil)
		return
	}
	var req requests.UpdateCommissionTierRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	res, err := c.commissionService.UpdateTier(ctx.Request.Context(), id, &req)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Tier updated successfully", res)
}

// DeleteTier godoc
// @Summary Delete commission tier
// @Description Admin only.
// @Tags admin-commission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tier ID"
// @Success 200 {object} responses.Response "Tier deleted"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/commission/tiers/{id} [delete]
func (c *CommissionController) DeleteTier(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid tier id", nil)
		return
	}
	if err := c.commissionService.DeleteTier(ctx.Request.Context(), id); err != nil {
		response.InternalServerError(ctx, "Failed to delete tier")
		return
	}
	response.OK(ctx, "Tier deleted successfully", nil)
}

// SetOverride godoc
// @Summary Set organizer commission override
// @Description Admin only. Rate bps = -1 disables override.
// @Tags admin-commission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.SetOrganizerCommissionOverrideRequest true "Override"
// @Success 200 {object} responses.OrganizerCommissionOverrideResponse "Override set"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/commission/overrides [post]
func (c *CommissionController) SetOverride(ctx *gin.Context) {
	var req requests.SetOrganizerCommissionOverrideRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID := ctx.MustGet("user_id").(uuid.UUID)
	res, err := c.commissionService.SetOverride(ctx.Request.Context(), actorID, &req)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Override set successfully", res)
}

// GetOverride godoc
// @Summary Get organizer commission override
// @Description Admin only.
// @Tags admin-commission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organizer_id path string true "Organizer ID"
// @Success 200 {object} responses.OrganizerCommissionOverrideResponse "Override"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/commission/overrides/{organizer_id} [get]
func (c *CommissionController) GetOverride(ctx *gin.Context) {
	organizerID, err := uuid.Parse(ctx.Param("organizer_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid organizer id", nil)
		return
	}
	res, err := c.commissionService.GetOverride(ctx.Request.Context(), organizerID)
	if err != nil {
		response.NotFound(ctx, "No override found")
		return
	}
	response.OK(ctx, "Override retrieved successfully", res)
}

// DeleteOverride godoc
// @Summary Delete organizer commission override
// @Description Admin only.
// @Tags admin-commission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organizer_id path string true "Organizer ID"
// @Success 200 {object} responses.Response "Override deleted"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/commission/overrides/{organizer_id} [delete]
func (c *CommissionController) DeleteOverride(ctx *gin.Context) {
	organizerID, err := uuid.Parse(ctx.Param("organizer_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid organizer id", nil)
		return
	}
	if err := c.commissionService.DeleteOverride(ctx.Request.Context(), organizerID); err != nil {
		response.InternalServerError(ctx, "Failed to delete override")
		return
	}
	response.OK(ctx, "Override deleted successfully", nil)
}
