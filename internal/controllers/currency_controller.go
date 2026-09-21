package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type CurrencyController struct {
	currencyService interfaces.CurrencyService
}

func NewCurrencyController(cs interfaces.CurrencyService) *CurrencyController {
	return &CurrencyController{currencyService: cs}
}

// CreateCurrency godoc
// @Summary Create a currency
// @Description Admin only. Adds a new currency and optionally enables it.
// @Tags admin-currencies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateCurrencyRequest true "Currency"
// @Success 201 {object} responses.CurrencyResponse "Currency created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 409 {object} responses.ErrorResponse "Currency exists"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/currencies [post]
func (c *CurrencyController) Create(ctx *gin.Context) {
	var req requests.CreateCurrencyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID := ctx.MustGet("user_id").(uuid.UUID)
	res, err := c.currencyService.Create(ctx.Request.Context(), actorID, &req)
	if err != nil {
		logger.WithRequest(ctx).Error("CreateCurrency failed: ", err)
		response.Conflict(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Currency created successfully", res)
}

// ListCurrencies godoc
// @Summary List currencies (admin)
// @Description Admin only. Lists all currencies, optionally filtered to enabled only.
// @Tags admin-currencies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param enabled query bool false "Only enabled currencies"
// @Success 200 {array} responses.CurrencyResponse "Currencies retrieved"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/currencies [get]
func (c *CurrencyController) ListAdmin(ctx *gin.Context) {
	onlyEnabled, _ := strconv.ParseBool(ctx.DefaultQuery("enabled", "false"))
	list, err := c.currencyService.List(ctx.Request.Context(), onlyEnabled)
	if err != nil {
		response.InternalServerError(ctx, "Failed to list currencies")
		return
	}
	response.OK(ctx, "Currencies retrieved successfully", list)
}

// ListEnabledCurrencies godoc
// @Summary List enabled currencies (public)
// @Description Public list of currencies enabled by the platform.
// @Tags currencies
// @Accept json
// @Produce json
// @Success 200 {array} responses.CurrencyResponse "Enabled currencies"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /currencies [get]
func (c *CurrencyController) ListEnabled(ctx *gin.Context) {
	list, err := c.currencyService.List(ctx.Request.Context(), true)
	if err != nil {
		response.InternalServerError(ctx, "Failed to list currencies")
		return
	}
	response.OK(ctx, "Currencies retrieved successfully", list)
}

// UpdateCurrency godoc
// @Summary Update a currency
// @Description Admin only.
// @Tags admin-currencies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Currency code"
// @Param request body requests.UpdateCurrencyRequest true "Update"
// @Success 200 {object} responses.CurrencyResponse "Currency updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/currencies/{code} [put]
func (c *CurrencyController) Update(ctx *gin.Context) {
	code := ctx.Param("code")
	var req requests.UpdateCurrencyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	res, err := c.currencyService.Update(ctx.Request.Context(), code, &req)
	if err != nil {
		if err.Error() == "currency not found" {
			response.NotFound(ctx, err.Error())
			return
		}
		response.InternalServerError(ctx, "Failed to update currency")
		return
	}
	response.OK(ctx, "Currency updated successfully", res)
}

// SetDefaultCurrency godoc
// @Summary Set default currency
// @Description Admin only. Sets the platform default currency.
// @Tags admin-currencies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Currency code"
// @Success 200 {object} responses.Response "Default updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/currencies/{code}/default [post]
func (c *CurrencyController) SetDefault(ctx *gin.Context) {
	code := ctx.Param("code")
	if err := c.currencyService.SetDefault(ctx.Request.Context(), code); err != nil {
		response.InternalServerError(ctx, "Failed to set default currency")
		return
	}
	response.OK(ctx, "Default currency updated", nil)
}

// RequestCurrency godoc
// @Summary Request a currency
// @Description Event Manager or Vendor requests a new currency to be enabled.
// @Tags currency-requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.RequestCurrencyRequest true "Request"
// @Success 201 {object} responses.CurrencyRequestResponse "Request submitted"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 409 {object} responses.ErrorResponse "Already exists"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /currency-requests [post]
func (c *CurrencyController) RequestCurrency(ctx *gin.Context) {
	var req requests.RequestCurrencyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID := ctx.MustGet("user_id").(uuid.UUID)
	res, err := c.currencyService.RequestCurrency(ctx.Request.Context(), actorID, &req)
	if err != nil {
		response.Conflict(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Currency request submitted", res)
}

// ListMyCurrencyRequests godoc
// @Summary List my currency requests
// @Description List current user's currency requests.
// @Tags currency-requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.CurrencyRequestResponse "My requests"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /currency-requests/me [get]
func (c *CurrencyController) ListMyRequests(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	actorID := ctx.MustGet("user_id").(uuid.UUID)
	list, total, err := c.currencyService.ListMyRequests(ctx.Request.Context(), actorID, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, "Failed to list requests")
		return
	}
	response.Paginated(ctx, "Currency requests retrieved successfully", list, page, perPage, total)
}

// ListAllCurrencyRequests godoc
// @Summary List all currency requests
// @Description Admin only.
// @Tags admin-currency-requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "pending | approved | rejected"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.CurrencyRequestResponse "Requests"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/currency-requests [get]
func (c *CurrencyController) ListAllRequests(ctx *gin.Context) {
	status := ctx.Query("status")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	list, total, err := c.currencyService.ListRequests(ctx.Request.Context(), status, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, "Failed to list requests")
		return
	}
	response.Paginated(ctx, "Currency requests retrieved successfully", list, page, perPage, total)
}

// ReviewCurrencyRequest godoc
// @Summary Review a currency request
// @Description Admin only. Approve or reject a currency request.
// @Tags admin-currency-requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Request ID"
// @Param request body requests.ReviewCurrencyRequestRequest true "Decision"
// @Success 200 {object} responses.CurrencyRequestResponse "Reviewed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/currency-requests/{id}/review [post]
func (c *CurrencyController) ReviewRequest(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid request id", nil)
		return
	}
	var req requests.ReviewCurrencyRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID := ctx.MustGet("user_id").(uuid.UUID)
	res, err := c.currencyService.ReviewRequest(ctx.Request.Context(), actorID, id, &req)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Currency request reviewed", res)
}
