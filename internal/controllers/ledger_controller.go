package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/response"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type LedgerController struct {
	service serviceInterfaces.LedgerService
}

func NewLedgerController(s serviceInterfaces.LedgerService) *LedgerController {
	return &LedgerController{service: s}
}

// GetMyBalance godoc
// @Summary Get my balance
// @Description Organizer balance: escrow held / available / paid out.
// @Tags ledger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param currency query string false "Currency code" default(NGN)
// @Success 200 {object} responses.BalanceResponse "Balance"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /ledger/balance [get]
func (c *LedgerController) MyBalance(ctx *gin.Context) {
	userID, _ := actorIDFromCtx(ctx)
	currency := ctx.DefaultQuery("currency", "NGN")
	res, err := c.service.GetBalance(ctx.Request.Context(), userID, currency)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Balance retrieved", res)
}

// ListMyEntries godoc
// @Summary List my ledger entries
// @Tags ledger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.LedgerEntryResponse "Entries"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /ledger/entries [get]
func (c *LedgerController) ListMine(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	userID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.ListEntries(ctx.Request.Context(), userID, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Ledger entries retrieved", list, page, perPage, total)
}

// GetOrganizerBalance godoc
// @Summary Get an organizer's balance (admin)
// @Description Admin only.
// @Tags admin-ledger
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organizer_id path string true "Organizer ID"
// @Param currency query string false "Currency code" default(NGN)
// @Success 200 {object} responses.BalanceResponse "Balance"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/ledger/{organizer_id}/balance [get]
func (c *LedgerController) OrganizerBalance(ctx *gin.Context) {
	orgID, err := uuid.Parse(ctx.Param("organizer_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid organizer id", nil)
		return
	}
	currency := ctx.DefaultQuery("currency", "NGN")
	res, err := c.service.GetBalance(ctx.Request.Context(), orgID, currency)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Balance retrieved", res)
}
