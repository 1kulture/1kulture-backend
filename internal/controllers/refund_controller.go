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

type RefundController struct {
	service serviceInterfaces.RefundService
}

func NewRefundController(s serviceInterfaces.RefundService) *RefundController {
	return &RefundController{service: s}
}

// RequestRefund godoc
// @Summary Request a refund
// @Description Buyer requests a refund for specific tickets.
// @Tags refunds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body requests.RefundRequestPayload true "Refund"
// @Success 201 {object} responses.RefundResponse "Requested"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /orders/{id}/refunds [post]
func (c *RefundController) Request(ctx *gin.Context) {
	orderID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid order id", nil)
		return
	}
	var req requests.RefundRequestPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.RequestRefund(ctx.Request.Context(), actorID, orderID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Refund requested", res)
}

// DecideRefund godoc
// @Summary Approve or reject a refund
// @Description Organizer/co-organizer only.
// @Tags refunds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Refund ID"
// @Param request body requests.RefundDecisionRequest true "Decision"
// @Success 200 {object} responses.RefundResponse "Reviewed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /refunds/{id}/decide [post]
func (c *RefundController) Decide(ctx *gin.Context) {
	refundID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid refund id", nil)
		return
	}
	var req requests.RefundDecisionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.DecideRefund(ctx.Request.Context(), actorID, refundID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Refund reviewed", res)
}

// GetRefund godoc
// @Summary Get a refund
// @Tags refunds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Refund ID"
// @Success 200 {object} responses.RefundResponse "Refund"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /refunds/{id} [get]
func (c *RefundController) Get(ctx *gin.Context) {
	refundID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid refund id", nil)
		return
	}
	viewerID, _ := actorIDFromCtx(ctx)
	res, err := c.service.GetRefund(ctx.Request.Context(), viewerID, refundID)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Refund retrieved", res)
}

// ListOrderRefunds godoc
// @Summary List refunds for an order
// @Tags refunds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {array} responses.RefundResponse "Refunds"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /orders/{id}/refunds [get]
func (c *RefundController) ListByOrder(ctx *gin.Context) {
	orderID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid order id", nil)
		return
	}
	viewerID, _ := actorIDFromCtx(ctx)
	res, err := c.service.ListByOrder(ctx.Request.Context(), viewerID, orderID)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Refunds retrieved", res)
}

// ListRefundsAdmin godoc
// @Summary List refunds by status (admin)
// @Description Admin only.
// @Tags admin-refunds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Status"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.RefundResponse "Refunds"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/refunds [get]
func (c *RefundController) ListAdmin(ctx *gin.Context) {
	status := ctx.Query("status")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	list, total, err := c.service.ListByStatus(ctx.Request.Context(), status, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Refunds retrieved", list, page, perPage, total)
}
