package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/1kulture/1kulture-backend/internal/requests"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type InterestController struct {
	service serviceInterfaces.InterestService
}

func NewInterestController(s serviceInterfaces.InterestService) *InterestController {
	return &InterestController{service: s}
}

// ListCategories godoc
// @Summary List event categories
// @Description Returns all active categories. Used to render the interest picker.
// @Tags interests
// @Accept json
// @Produce json
// @Success 200 {array} responses.EventCategoryResponse "Categories"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /categories [get]
func (c *InterestController) ListCategories(ctx *gin.Context) {
	res, err := c.service.ListCategories(ctx.Request.Context())
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Categories retrieved", res)
}

// GetMyInterests godoc
// @Summary Get my interests
// @Description Returns the current user's selected categories.
// @Tags interests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.UserInterestsResponse "My interests"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /users/me/interests [get]
func (c *InterestController) GetMine(ctx *gin.Context) {
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.GetMyInterests(ctx.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Interests retrieved", res)
}

// UpdateMyInterests godoc
// @Summary Update my interests
// @Description Replaces the current user's interests with the given set of category slugs.
// @Description Send an empty array to clear all interests.
// @Tags interests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.UpdateInterestsRequest true "Interests"
// @Success 200 {object} responses.UserInterestsResponse "Updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /users/me/interests [put]
func (c *InterestController) UpdateMine(ctx *gin.Context) {
	var req requests.UpdateInterestsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.UpdateMyInterests(ctx.Request.Context(), userID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Interests updated", res)
}