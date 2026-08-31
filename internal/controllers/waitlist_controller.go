package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"
)

type WaitlistController struct {
	waitlistService interfaces.WaitlistService
}

func NewWaitlistController(waitlistService interfaces.WaitlistService) *WaitlistController {
	return &WaitlistController{
		waitlistService: waitlistService,
	}
}

// AddToWaitlist godoc
// @Summary Add to waitlist
// @Description Add email and category to waitlist
// @Tags waitlist
// @Accept json
// @Produce json
// @Param request body requests.WaitlistRequest true "Waitlist request"
// @Success 201 {object} Response "Added to waitlist"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /waitlist [post]
func (c *WaitlistController) AddToWaitlist(ctx *gin.Context) {
	var req requests.WaitlistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}

	if errors := validator.Struct(req); errors != nil {
		response.ValidationError(ctx, errors)
		return
	}

	if err := c.waitlistService.AddToWaitlist(ctx.Request.Context(), &req); err != nil {
		logger.WithRequest(ctx).Error("AddToWaitlist failed: ", err)
		response.InternalServerError(ctx, "Failed to add to waitlist")
		return
	}

	response.Created(ctx, "Successfully added to waitlist", nil)
}
