package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/requests"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type OrderController struct {
	service serviceInterfaces.OrderService
}

func NewOrderController(s serviceInterfaces.OrderService) *OrderController {
	return &OrderController{service: s}
}

// CreateOrder godoc
// @Summary Create an order
// @Description Create a pending order. Reserves stock and applies any promo code.
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateOrderRequest true "Order"
// @Success 201 {object} responses.OrderResponse "Order created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /orders [post]
func (c *OrderController) Create(ctx *gin.Context) {
	var req requests.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.CreateOrder(ctx.Request.Context(), userID, &req)
	if err != nil {
		logger.WithRequest(ctx).Error("CreateOrder: ", err)
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Order created", res)
}

// InitializePayment godoc
// @Summary Initialize payment for an order
// @Description Starts a Paystack session and returns the authorization URL.
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body requests.InitializePaymentRequest true "Init"
// @Success 200 {object} responses.InitializePaymentResponse "Payment initialized"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /orders/{id}/pay [post]
func (c *OrderController) InitializePayment(ctx *gin.Context) {
	orderID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid order id", nil)
		return
	}
	var req requests.InitializePaymentRequest
	_ = ctx.ShouldBindJSON(&req) // body optional
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.InitializePayment(ctx.Request.Context(), userID, orderID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Payment initialized", res)
}

// GetOrder godoc
// @Summary Get an order
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} responses.OrderResponse "Order"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /orders/{id} [get]
func (c *OrderController) Get(ctx *gin.Context) {
	orderID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid order id", nil)
		return
	}
	viewerID, _ := actorIDFromCtx(ctx)
	res, err := c.service.GetOrder(ctx.Request.Context(), viewerID, orderID)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Order retrieved", res)
}

// ListMyOrders godoc
// @Summary List my orders
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.OrderResponse "Orders"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /orders/me [get]
func (c *OrderController) ListMine(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	userID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.ListMyOrders(ctx.Request.Context(), userID, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Orders retrieved", list, page, perPage, total)
}

// ListEventOrders godoc
// @Summary List orders for an event
// @Description Organizer/co-organizer only.
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.OrderResponse "Orders"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/orders [get]
func (c *OrderController) ListEventOrders(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	actorID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.ListEventOrders(ctx.Request.Context(), actorID, eventID, page, perPage)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Forbidden(ctx, "You cannot view orders for this event")
			return
		}
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Orders retrieved", list, page, perPage, total)
}

// PaystackWebhook godoc
// @Summary Paystack webhook
// @Description Called by Paystack. Do NOT call manually.
// @Tags webhooks
// @Accept json
// @Produce json
// @Param X-Paystack-Signature header string true "Signature"
// @Success 200 {object} responses.Response "OK"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Router /webhooks/paystack [post]
func (c *OrderController) PaystackWebhook(ctx *gin.Context) {
	body, err := ctx.GetRawData()
	if err != nil {
		response.BadRequest(ctx, "Failed to read body", nil)
		return
	}
	sig := ctx.GetHeader("X-Paystack-Signature")
	if err := c.service.HandleWebhook(ctx.Request.Context(), "paystack", body, sig); err != nil {
		logger.WithRequest(ctx).Error("Webhook error: ", err)
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Webhook processed", nil)
}
