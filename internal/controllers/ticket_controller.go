package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/response"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type TicketController struct {
	service serviceInterfaces.TicketService
}

func NewTicketController(s serviceInterfaces.TicketService) *TicketController {
	return &TicketController{service: s}
}

// GetTicket godoc
// @Summary Get a ticket by ID
// @Description Buyer or event organizer can view.
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket ID"
// @Success 200 {object} responses.TicketWithQRResponse "Ticket"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /tickets/{id} [get]
func (c *TicketController) Get(ctx *gin.Context) {
	tid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid ticket id", nil)
		return
	}
	viewerID, _ := actorIDFromCtx(ctx)
	res, err := c.service.GetTicket(ctx.Request.Context(), viewerID, tid)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Ticket retrieved", res)
}

// GetTicketByCode godoc
// @Summary Get a ticket by code
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Ticket code"
// @Success 200 {object} responses.TicketWithQRResponse "Ticket"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /tickets/code/{code} [get]
func (c *TicketController) GetByCode(ctx *gin.Context) {
	code := ctx.Param("code")
	viewerID, _ := actorIDFromCtx(ctx)
	res, err := c.service.GetTicketByCode(ctx.Request.Context(), viewerID, code)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Ticket retrieved", res)
}

// ListMyTickets godoc
// @Summary List my tickets
// @Description All tickets owned by the authenticated user.
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.TicketResponse "Tickets"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /tickets/me [get]
func (c *TicketController) ListMine(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	userID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.ListMyTickets(ctx.Request.Context(), userID, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Tickets retrieved", list, page, perPage, total)
}

// ListEventTickets godoc
// @Summary List tickets sold for an event
// @Description Organizer/co-organizer only.
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.TicketResponse "Tickets"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/tickets [get]
func (c *TicketController) ListEventTickets(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	actorID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.ListEventTickets(ctx.Request.Context(), actorID, eventID, page, perPage)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Forbidden(ctx, "You cannot view tickets for this event")
			return
		}
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Tickets retrieved", list, page, perPage, total)
}
