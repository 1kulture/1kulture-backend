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

type CheckInController struct {
	service serviceInterfaces.CheckInService
}

func NewCheckInController(s serviceInterfaces.CheckInService) *CheckInController {
	return &CheckInController{service: s}
}

// ScanQR godoc
// @Summary Scan ticket QR
// @Description Staff/organizer scans a ticket via its QR payload.
// @Tags check-in
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.CheckInRequest true "Scan"
// @Success 200 {object} responses.CheckInResponse "Checked in"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/checkin/scan [post]
func (c *CheckInController) ScanQR(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.CheckInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.ScanByQR(ctx.Request.Context(), actorID, eventID, &req, ctx.ClientIP())
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Ticket checked in", res)
}

// ManualCheckIn godoc
// @Summary Manual check-in
// @Description Staff/organizer performs a manual check-in using the ticket code.
// @Tags check-in
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.ManualCheckInRequest true "Manual"
// @Success 200 {object} responses.CheckInResponse "Checked in"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/checkin/manual [post]
func (c *CheckInController) Manual(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.ManualCheckInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.ManualCheckIn(ctx.Request.Context(), actorID, eventID, &req, ctx.ClientIP())
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Ticket checked in", res)
}

// ListCheckIns godoc
// @Summary List check-ins for an event
// @Tags check-in
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.CheckInRecordResponse "Check-ins"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/checkin [get]
func (c *CheckInController) List(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "50"))
	actorID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.ListCheckIns(ctx.Request.Context(), actorID, eventID, page, perPage)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Paginated(ctx, "Check-ins retrieved", list, page, perPage, total)
}

// Stats godoc
// @Summary Check-in stats
// @Tags check-in
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} responses.CheckInStatsResponse "Stats"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/checkin/stats [get]
func (c *CheckInController) Stats(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Stats(ctx.Request.Context(), actorID, eventID)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Stats retrieved", res)
}
