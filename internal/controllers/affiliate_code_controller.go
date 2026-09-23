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

type AffiliateCodeController struct {
	service serviceInterfaces.AffiliateCodeService
}

func NewAffiliateCodeController(s serviceInterfaces.AffiliateCodeService) *AffiliateCodeController {
	return &AffiliateCodeController{service: s}
}

// CreateAffiliateCode godoc
// @Summary Create an affiliate code
// @Description Organizer/co-organizer only. Creates a promo code tied to a partnership.
// @Tags affiliate-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partnership ID"
// @Param request body requests.CreateAffiliateCodeRequest true "Code"
// @Success 201 {object} responses.AffiliateCodeResponse "Created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 409 {object} responses.ErrorResponse "Code exists"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/{id}/affiliate-codes [post]
func (c *AffiliateCodeController) Create(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid partnership id", nil)
		return
	}
	var req requests.CreateAffiliateCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Create(ctx.Request.Context(), actorID, id, &req)
	if err != nil {
		response.Conflict(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Affiliate code created", res)
}

// ListAffiliateCodes godoc
// @Summary List affiliate codes for a partnership
// @Tags affiliate-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partnership ID"
// @Success 200 {array} responses.AffiliateCodeResponse "Codes"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/{id}/affiliate-codes [get]
func (c *AffiliateCodeController) List(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid partnership id", nil)
		return
	}
	viewerID, _ := actorIDFromCtx(ctx)
	res, err := c.service.List(ctx.Request.Context(), viewerID, id)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Affiliate codes retrieved", res)
}
