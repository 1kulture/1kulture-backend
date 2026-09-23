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

type PartnershipConfigController struct {
	service serviceInterfaces.PartnershipConfigService
}

func NewPartnershipConfigController(s serviceInterfaces.PartnershipConfigService) *PartnershipConfigController {
	return &PartnershipConfigController{service: s}
}

// UpsertPartnershipConfig godoc
// @Summary Configure event partnership mode
// @Description Organizer/co-organizer only. Creates or updates the partnership config for an event.
// @Tags event-partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.EventPartnershipConfigRequest true "Config"
// @Success 200 {object} responses.EventPartnershipConfigResponse "Config saved"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/partnership-config [put]
func (c *PartnershipConfigController) Upsert(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.EventPartnershipConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.UpsertConfig(ctx.Request.Context(), actorID, eventID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Partnership config saved", res)
}

// GetPartnershipConfig godoc
// @Summary Get event partnership config
// @Description Public for published events; else requires organizer access.
// @Tags event-partnerships
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} responses.EventPartnershipConfigResponse "Config"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/partnership-config [get]
func (c *PartnershipConfigController) Get(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	res, err := c.service.GetConfig(ctx.Request.Context(), eventID, viewerIDFromCtx(ctx))
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Partnership config retrieved", res)
}

// AddOpportunity godoc
// @Summary Add partnership opportunity
// @Description Organizer/co-organizer only.
// @Tags event-partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.OpportunityCreateRequest true "Opportunity"
// @Success 201 {object} responses.OpportunityResponse "Created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/partnership-opportunities [post]
func (c *PartnershipConfigController) AddOpportunity(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.OpportunityCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.AddOpportunity(ctx.Request.Context(), actorID, eventID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Opportunity added", res)
}

// UpdateOpportunity godoc
// @Summary Update partnership opportunity
// @Description Organizer/co-organizer only.
// @Tags event-partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param opp_id path string true "Opportunity ID"
// @Param request body requests.OpportunityUpdateRequest true "Update"
// @Success 200 {object} responses.OpportunityResponse "Updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/partnership-opportunities/{opp_id} [put]
func (c *PartnershipConfigController) UpdateOpportunity(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	oppID, err := uuid.Parse(ctx.Param("opp_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid opportunity id", nil)
		return
	}
	var req requests.OpportunityUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.UpdateOpportunity(ctx.Request.Context(), actorID, eventID, oppID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Opportunity updated", res)
}

// DeleteOpportunity godoc
// @Summary Delete partnership opportunity
// @Description Organizer/co-organizer only.
// @Tags event-partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param opp_id path string true "Opportunity ID"
// @Success 200 {object} responses.Response "Deleted"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/partnership-opportunities/{opp_id} [delete]
func (c *PartnershipConfigController) DeleteOpportunity(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	oppID, err := uuid.Parse(ctx.Param("opp_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid opportunity id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	if err := c.service.DeleteOpportunity(ctx.Request.Context(), actorID, eventID, oppID); err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Opportunity deleted", nil)
}

// ListOpportunities godoc
// @Summary List partnership opportunities for an event
// @Tags event-partnerships
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param active_only query bool false "Only active"
// @Success 200 {array} responses.OpportunityResponse "Opportunities"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/partnership-opportunities [get]
func (c *PartnershipConfigController) ListOpportunities(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	onlyActive := ctx.DefaultQuery("active_only", "false") == "true"
	list, err := c.service.ListOpportunities(ctx.Request.Context(), eventID, onlyActive)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Opportunities retrieved", list)
}
