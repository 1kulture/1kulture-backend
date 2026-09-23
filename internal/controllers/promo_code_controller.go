package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/requests"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type PromoCodeController struct {
	service serviceInterfaces.PromoCodeService
}

func NewPromoCodeController(s serviceInterfaces.PromoCodeService) *PromoCodeController {
	return &PromoCodeController{service: s}
}

// CreatePromoCode godoc
// @Summary Create a promo code
// @Description Organizer/co-organizer only.
// @Tags promo-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.PromoCodeCreateRequest true "Promo"
// @Success 201 {object} responses.PromoCodeResponse "Created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/promo-codes [post]
func (c *PromoCodeController) Create(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.PromoCodeCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Create(ctx.Request.Context(), actorID, eventID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Promo code created", res)
}

// ListPromoCodes godoc
// @Summary List promo codes for an event
// @Description Organizer/co-organizer only.
// @Tags promo-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {array} responses.PromoCodeResponse "Promo codes"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/promo-codes [get]
func (c *PromoCodeController) List(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	list, err := c.service.List(ctx.Request.Context(), actorID, eventID)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Promo codes retrieved", list)
}

// UpdatePromoCode godoc
// @Summary Update a promo code
// @Description Organizer/co-organizer only.
// @Tags promo-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param promo_id path string true "Promo Code ID"
// @Param request body requests.PromoCodeUpdateRequest true "Update"
// @Success 200 {object} responses.PromoCodeResponse "Updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/promo-codes/{promo_id} [put]
func (c *PromoCodeController) Update(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	promoID, err := uuid.Parse(ctx.Param("promo_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid promo id", nil)
		return
	}
	var req requests.PromoCodeUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Update(ctx.Request.Context(), actorID, eventID, promoID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Promo code updated", res)
}

// DeletePromoCode godoc
// @Summary Delete a promo code
// @Description Organizer/co-organizer only.
// @Tags promo-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param promo_id path string true "Promo Code ID"
// @Success 200 {object} responses.Response "Deleted"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/promo-codes/{promo_id} [delete]
func (c *PromoCodeController) Delete(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	promoID, err := uuid.Parse(ctx.Param("promo_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid promo id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	if err := c.service.Delete(ctx.Request.Context(), actorID, eventID, promoID); err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Promo code deleted", nil)
}

// ValidatePromoCode godoc
// @Summary Validate a promo code
// @Description Public preview of a promo code's effect on a subtotal.
// @Tags promo-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.ValidatePromoRequest true "Validate"
// @Success 200 {object} responses.PromoValidationResponse "Promo preview"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /promo-codes/validate [post]
func (c *PromoCodeController) Validate(ctx *gin.Context) {
	var req requests.ValidatePromoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Validate(ctx.Request.Context(), userID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Promo code applied", res)
}
