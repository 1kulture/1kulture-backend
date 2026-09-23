package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/requests"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type TicketTypeController struct {
	service serviceInterfaces.TicketTypeService
}

func NewTicketTypeController(s serviceInterfaces.TicketTypeService) *TicketTypeController {
	return &TicketTypeController{service: s}
}

// CreateTicketType godoc
// @Summary Create a ticket type
// @Description Organizer/co-organizer only.
// @Tags ticket-types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.TicketTypeCreateRequest true "Ticket type"
// @Success 201 {object} responses.TicketTypeResponse "Created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/ticket-types [post]
func (c *TicketTypeController) Create(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.TicketTypeCreateRequest
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
		logger.WithRequest(ctx).Error("CreateTicketType: ", err)
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Ticket type created", res)
}

// ListTicketTypes godoc
// @Summary List ticket types for an event
// @Tags ticket-types
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {array} responses.TicketTypeResponse "Ticket types"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/ticket-types [get]
func (c *TicketTypeController) List(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	list, err := c.service.List(ctx.Request.Context(), eventID, viewerIDFromCtx(ctx))
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Ticket types retrieved", list)
}

// UpdateTicketType godoc
// @Summary Update a ticket type
// @Description Organizer/co-organizer only.
// @Tags ticket-types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param tt_id path string true "Ticket Type ID"
// @Param request body requests.TicketTypeUpdateRequest true "Update"
// @Success 200 {object} responses.TicketTypeResponse "Updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/ticket-types/{tt_id} [put]
func (c *TicketTypeController) Update(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	ttID, err := uuid.Parse(ctx.Param("tt_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid ticket type id", nil)
		return
	}
	var req requests.TicketTypeUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Update(ctx.Request.Context(), actorID, eventID, ttID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Ticket type updated", res)
}

// DeleteTicketType godoc
// @Summary Delete a ticket type
// @Description Organizer/co-organizer only.
// @Tags ticket-types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param tt_id path string true "Ticket Type ID"
// @Success 200 {object} responses.Response "Deleted"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/ticket-types/{tt_id} [delete]
func (c *TicketTypeController) Delete(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	ttID, err := uuid.Parse(ctx.Param("tt_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid ticket type id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	if err := c.service.Delete(ctx.Request.Context(), actorID, eventID, ttID); err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Ticket type deleted", nil)
}
