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

type PartnershipMetricController struct {
	service serviceInterfaces.PartnershipMetricService
}

func NewPartnershipMetricController(s serviceInterfaces.PartnershipMetricService) *PartnershipMetricController {
	return &PartnershipMetricController{service: s}
}

// AddMetric godoc
// @Summary Report a partnership metric
// @Description Either side (organizer or brand) can report.
// @Tags partnership-metrics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partnership ID"
// @Param request body requests.PartnershipMetricRequest true "Metric"
// @Success 201 {object} responses.PartnershipMetricResponse "Added"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/{id}/metrics [post]
func (c *PartnershipMetricController) Add(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid partnership id", nil)
		return
	}
	var req requests.PartnershipMetricRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Add(ctx.Request.Context(), actorID, id, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Metric added", res)
}

// ListMetrics godoc
// @Summary List partnership metrics
// @Tags partnership-metrics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partnership ID"
// @Success 200 {array} responses.PartnershipMetricResponse "Metrics"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/{id}/metrics [get]
func (c *PartnershipMetricController) List(ctx *gin.Context) {
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
	response.OK(ctx, "Metrics retrieved", res)
}

// DeleteMetric godoc
// @Summary Delete a metric
// @Tags partnership-metrics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Metric ID"
// @Success 200 {object} responses.Response "Deleted"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnership-metrics/{id} [delete]
func (c *PartnershipMetricController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid metric id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	if err := c.service.Delete(ctx.Request.Context(), actorID, id); err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Metric deleted", nil)
}
