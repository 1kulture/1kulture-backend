package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type BrandDiscoveryController struct {
	service serviceInterfaces.BrandDiscoveryService
}

func NewBrandDiscoveryController(s serviceInterfaces.BrandDiscoveryService) *BrandDiscoveryController {
	return &BrandDiscoveryController{service: s}
}

// DiscoverEvents godoc
// @Summary Discover events open for partnerships
// @Description Brand-facing discovery with partnership-specific filters.
// @Tags brand-discovery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category query string false "Category"
// @Param city query string false "City"
// @Param country query string false "Country"
// @Param partnership_type query string false "Partnership type (cash_sponsorship, product_sponsorship, ...)"
// @Param min_audience query int false "Minimum expected attendees"
// @Param max_audience query int false "Maximum expected attendees"
// @Param start_from query string false "RFC3339 start filter"
// @Param start_to query string false "RFC3339 end filter"
// @Param search query string false "Search title/summary"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.EventSummaryResponse "Events"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands/discover [get]
func (c *BrandDiscoveryController) DiscoverEvents(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	minAud, _ := strconv.Atoi(ctx.DefaultQuery("min_audience", "0"))
	maxAud, _ := strconv.Atoi(ctx.DefaultQuery("max_audience", "0"))

	filter := serviceInterfaces.BrandDiscoveryFilter{
		Category:        ctx.Query("category"),
		City:            ctx.Query("city"),
		Country:         ctx.Query("country"),
		PartnershipType: ctx.Query("partnership_type"),
		MinAudience:     minAud,
		MaxAudience:     maxAud,
		StartFromISO:    ctx.Query("start_from"),
		StartToISO:      ctx.Query("start_to"),
		Search:          ctx.Query("search"),
		Page:            page,
		PerPage:         perPage,
	}
	userID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.DiscoverEvents(ctx.Request.Context(), userID, filter)
	if err != nil {
		logger.WithRequest(ctx).Error("DiscoverEvents: ", err)
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Events retrieved", list, page, perPage, total)
}

// RecommendedEvents godoc
// @Summary Recommended events for the brand
// @Description Uses brand's preferred categories.
// @Tags brand-discovery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} responses.EventSummaryResponse "Events"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands/recommended [get]
func (c *BrandDiscoveryController) Recommended(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	userID, _ := actorIDFromCtx(ctx)
	list, err := c.service.Recommended(ctx.Request.Context(), userID, limit)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Recommended events retrieved", list)
}

// BrandDashboard godoc
// @Summary Brand dashboard
// @Description Active partnerships, pending requests, available events.
// @Tags brand-discovery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.BrandDashboardResponse "Dashboard"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands/dashboard [get]
func (c *BrandDiscoveryController) Dashboard(ctx *gin.Context) {
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Dashboard(ctx.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Dashboard retrieved", res)
}

// BrandEventDetail godoc
// @Summary Get event detail (with partnership config) for a brand
// @Tags brand-discovery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} responses.Response "Event + config"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands/events/{id} [get]
func (c *BrandDiscoveryController) EventDetail(ctx *gin.Context) {
	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	event, cfg, err := c.service.GetEventDetail(ctx.Request.Context(), userID, eventID)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Event retrieved", gin.H{
		"event":              event,
		"partnership_config": cfg,
	})
}
