package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type PartnershipController struct {
	service serviceInterfaces.PartnershipService
}

func NewPartnershipController(s serviceInterfaces.PartnershipService) *PartnershipController {
	return &PartnershipController{service: s}
}

// CreatePartnershipRequest godoc
// @Summary Request a partnership
// @Description Brand requests to partner with an event.
// @Tags partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.PartnershipRequestCreate true "Request"
// @Success 201 {object} responses.PartnershipRequestResponse "Created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 409 {object} responses.ErrorResponse "Already requested"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/partnership-requests [post]
func (c *PartnershipController) Create(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.PartnershipRequestCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.CreateRequest(ctx.Request.Context(), userID, eventID, &req)
	if err != nil {
		logger.WithRequest(ctx).Error("CreatePartnershipRequest: ", err)
		response.Conflict(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Partnership requested", res)
}

// ListMyBrandRequests godoc
// @Summary List my partnership requests (brand)
// @Description Requests made by the authenticated brand.
// @Tags partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.PartnershipRequestResponse "Requests"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands/partnerships [get]
func (c *PartnershipController) ListMine(ctx *gin.Context) {
	status := ctx.Query("status")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	userID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.ListMyBrandRequests(ctx.Request.Context(), userID, status, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Partnership requests retrieved", list, page, perPage, total)
}

// CancelPartnershipRequest godoc
// @Summary Cancel a partnership request (brand)
// @Tags partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partnership Request ID"
// @Param request body requests.PartnershipDeclineRequest true "Reason"
// @Success 200 {object} responses.PartnershipRequestResponse "Cancelled"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/{id}/cancel [post]
func (c *PartnershipController) Cancel(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid partnership id", nil)
		return
	}
	var req requests.PartnershipDeclineRequest
	_ = ctx.ShouldBindJSON(&req)
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.CancelRequest(ctx.Request.Context(), userID, id, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Partnership request cancelled", res)
}

// ListEventRequests godoc
// @Summary List partnership requests for an event (organizer)
// @Description Organizer/co-organizer only.
// @Tags partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param status query string false "Filter by status"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.PartnershipRequestResponse "Requests"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/partnership-requests [get]
func (c *PartnershipController) ListEventRequests(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	status := ctx.Query("status")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	actorID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.ListEventRequests(ctx.Request.Context(), actorID, eventID, status, page, perPage)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Paginated(ctx, "Partnership requests retrieved", list, page, perPage, total)
}

// DecidePartnershipRequest godoc
// @Summary Approve or decline a partnership request (organizer)
// @Description Organizer/co-organizer only.
// @Tags partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partnership Request ID"
// @Param accept query bool true "Accept or decline"
// @Param request body requests.PartnershipDeclineRequest true "Reason (for decline)"
// @Success 200 {object} responses.PartnershipRequestResponse "Decided"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/{id}/decide [post]
func (c *PartnershipController) Decide(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid partnership id", nil)
		return
	}
	accept := ctx.DefaultQuery("accept", "false") == "true"
	var req requests.PartnershipDeclineRequest
	_ = ctx.ShouldBindJSON(&req)
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.DecideRequest(ctx.Request.Context(), actorID, id, accept, req.Reason)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Partnership decision recorded", res)
}

// ActivatePartnership godoc
// @Summary Mark a partnership as active
// @Description Organizer/co-organizer only.
// @Tags partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partnership Request ID"
// @Success 200 {object} responses.PartnershipRequestResponse "Activated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/{id}/activate [post]
func (c *PartnershipController) Activate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid partnership id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.MarkActivated(ctx.Request.Context(), actorID, id)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Partnership activated", res)
}

// CompletePartnership godoc
// @Summary Mark a partnership as completed
// @Description Organizer/co-organizer only.
// @Tags partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partnership Request ID"
// @Success 200 {object} responses.PartnershipRequestResponse "Completed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/{id}/complete [post]
func (c *PartnershipController) Complete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid partnership id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.service.MarkCompleted(ctx.Request.Context(), actorID, id)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Partnership completed", res)
}

// GetPartnership godoc
// @Summary Get a partnership
// @Tags partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Partnership Request ID"
// @Success 200 {object} responses.PartnershipRequestResponse "Partnership"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/{id} [get]
func (c *PartnershipController) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid partnership id", nil)
		return
	}
	viewerID, _ := actorIDFromCtx(ctx)
	res, err := c.service.GetRequest(ctx.Request.Context(), viewerID, id)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Partnership retrieved", res)
}

// PartnershipDashboard godoc
// @Summary Partnership dashboard
// @Description Organizer dashboard: counts by status. Pass event_id to scope to one event.
// @Tags partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param event_id query string false "Optional Event ID to scope"
// @Success 200 {object} responses.PartnershipDashboardResponse "Dashboard"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /partnerships/dashboard [get]
func (c *PartnershipController) Dashboard(ctx *gin.Context) {
	var eventID *uuid.UUID
	if v := ctx.Query("event_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.BadRequest(ctx, "Invalid event_id", nil)
			return
		}
		eventID = &id
	}
	viewerID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Dashboard(ctx.Request.Context(), viewerID, eventID)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Dashboard retrieved", res)
}

// AdminListPartnerships godoc
// @Summary List all partnerships (admin)
// @Description Admin only.
// @Tags admin-partnerships
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Status"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.PartnershipRequestResponse "Partnerships"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /admin/partnerships [get]
func (c *PartnershipController) AdminList(ctx *gin.Context) {
	status := ctx.Query("status")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	viewerID, _ := actorIDFromCtx(ctx)
	filter := interfaces.PartnershipRequestFilter{
		Status:  status,
		Page:    page,
		PerPage: perPage,
	}
	list, total, err := c.service.ListAll(ctx.Request.Context(), filter, viewerID)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Partnerships retrieved", list, page, perPage, total)
}
