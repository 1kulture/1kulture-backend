package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/requests"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type TicketTransferController struct {
	service serviceInterfaces.TicketTransferService
}

func NewTicketTransferController(s serviceInterfaces.TicketTransferService) *TicketTransferController {
	return &TicketTransferController{service: s}
}

// InitiateTransfer godoc
// @Summary Initiate ticket transfer
// @Description Owner sends a ticket to another email.
// @Tags ticket-transfers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket ID"
// @Param request body requests.InitiateTransferRequest true "Transfer"
// @Success 201 {object} responses.TicketTransferResponse "Initiated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /tickets/{id}/transfer [post]
func (c *TicketTransferController) Initiate(ctx *gin.Context) {
	ticketID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid ticket id", nil)
		return
	}
	var req requests.InitiateTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.InitiateTransfer(ctx.Request.Context(), actorID, ticketID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Transfer initiated", res)
}

// AcceptTransfer godoc
// @Summary Accept a ticket transfer
// @Tags ticket-transfers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.AcceptTransferRequest true "Accept"
// @Success 200 {object} responses.TicketResponse "Accepted"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /ticket-transfers/accept [post]
func (c *TicketTransferController) Accept(ctx *gin.Context) {
	var req requests.AcceptTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.AcceptTransfer(ctx.Request.Context(), actorID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Transfer accepted", res)
}

// DeclineTransfer godoc
// @Summary Decline a ticket transfer
// @Tags ticket-transfers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.DeclineTransferRequest true "Decline"
// @Success 200 {object} responses.TicketTransferResponse "Declined"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /ticket-transfers/decline [post]
func (c *TicketTransferController) Decline(ctx *gin.Context) {
	var req requests.DeclineTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.DeclineTransfer(ctx.Request.Context(), actorID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Transfer declined", res)
}

// CancelTransfer godoc
// @Summary Cancel a ticket transfer
// @Tags ticket-transfers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transfer ID"
// @Success 200 {object} responses.TicketTransferResponse "Cancelled"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /ticket-transfers/{id}/cancel [post]
func (c *TicketTransferController) Cancel(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid transfer id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.CancelTransfer(ctx.Request.Context(), actorID, id)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Transfer cancelled", res)
}

// GetTransferByToken godoc
// @Summary Get a ticket transfer by token
// @Description Public lookup — used by the recipient before accepting.
// @Tags ticket-transfers
// @Accept json
// @Produce json
// @Param token path string true "Transfer token"
// @Success 200 {object} responses.TicketTransferResponse "Transfer"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /ticket-transfers/token/{token} [get]
func (c *TicketTransferController) GetByToken(ctx *gin.Context) {
	token := ctx.Param("token")
	res, err := c.service.GetTransferByToken(ctx.Request.Context(), token)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Transfer retrieved", res)
}

// GetTransfer godoc
// @Summary Get a ticket transfer
// @Tags ticket-transfers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transfer ID"
// @Success 200 {object} responses.TicketTransferResponse "Transfer"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /ticket-transfers/{id} [get]
func (c *TicketTransferController) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid transfer id", nil)
		return
	}
	viewerID, _ := actorIDFromCtx(ctx)
	res, err := c.service.GetTransfer(ctx.Request.Context(), viewerID, id)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Transfer retrieved", res)
}

// ListMyTransfers godoc
// @Summary List my ticket transfers
// @Tags ticket-transfers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.TicketTransferResponse "Transfers"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /ticket-transfers/me [get]
func (c *TicketTransferController) ListMine(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	userID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.ListMyTransfers(ctx.Request.Context(), userID, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Transfers retrieved", list, page, perPage, total)
}
