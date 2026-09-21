package controllers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
	"github.com/1kulture/1kulture-backend/internal/utils/response"
	"github.com/1kulture/1kulture-backend/internal/utils/validator"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type EventController struct {
	eventService serviceInterfaces.EventService
}

func NewEventController(eventService serviceInterfaces.EventService) *EventController {
	return &EventController{eventService: eventService}
}

// viewerIDFromCtx returns the authenticated user's ID from context, or nil.
func viewerIDFromCtx(ctx *gin.Context) *uuid.UUID {
	v, ok := ctx.Get("user_id")
	if !ok {
		return nil
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}

// actorIDFromCtx returns the authenticated user's ID and true, or uuid.Nil and false.
func actorIDFromCtx(ctx *gin.Context) (uuid.UUID, bool) {
	v, ok := ctx.Get("user_id")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}
	return id, true
}

// ----------------------------------------------------------
// CRUD
// ----------------------------------------------------------

// CreateEvent godoc
// @Summary Create event
// @Description Create a draft event. Requires EventManager role.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.EventCreateRequest true "Event"
// @Success 201 {object} responses.EventResponse "Event created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events [post]
func (c *EventController) Create(ctx *gin.Context) {
	var req requests.EventCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, ok := actorIDFromCtx(ctx)
	if !ok {
		response.Unauthorized(ctx, "Unauthenticated", nil)
		return
	}
	res, err := c.eventService.Create(ctx.Request.Context(), actorID, &req)
	if err != nil {
		logger.WithRequest(ctx).Error("CreateEvent failed: ", err)
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Event created successfully", res)
}

// ListEvents godoc
// @Summary List events
// @Description Public listing. Supports filters.
// @Tags events
// @Accept json
// @Produce json
// @Param category query string false "Category"
// @Param city query string false "City"
// @Param country query string false "Country"
// @Param event_type query string false "in_person | virtual | hybrid"
// @Param search query string false "Search title/summary"
// @Param organizer_id query string false "Organizer ID"
// @Param status query string false "Status"
// @Param start_from query string false "RFC3339 start filter"
// @Param start_to query string false "RFC3339 end filter"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.EventSummaryResponse "Events"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events [get]
func (c *EventController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))

	filter := repoInterfaces.EventListFilter{
		Category:  ctx.Query("category"),
		City:      ctx.Query("city"),
		Country:   ctx.Query("country"),
		EventType: ctx.Query("event_type"),
		Search:    ctx.Query("search"),
		Status:    ctx.Query("status"),
		Page:      page,
		PerPage:   perPage,
	}
	if v := ctx.Query("organizer_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.BadRequest(ctx, "Invalid organizer_id", nil)
			return
		}
		filter.OrganizerID = &id
	}
	if v := ctx.Query("start_from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.BadRequest(ctx, "Invalid start_from", nil)
			return
		}
		filter.StartFrom = &t
	}
	if v := ctx.Query("start_to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.BadRequest(ctx, "Invalid start_to", nil)
			return
		}
		filter.StartTo = &t
	}

	list, total, err := c.eventService.List(ctx.Request.Context(), filter, viewerIDFromCtx(ctx))
	if err != nil {
		logger.WithRequest(ctx).Error("ListEvents failed: ", err)
		response.InternalServerError(ctx, "Failed to list events")
		return
	}
	response.Paginated(ctx, "Events retrieved successfully", list, page, perPage, total)
}

// GetEvent godoc
// @Summary Get event by ID
// @Description Public if published; otherwise requires organizer access.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} responses.EventResponse "Event"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id} [get]
func (c *EventController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	res, err := c.eventService.GetByID(ctx.Request.Context(), id, viewerIDFromCtx(ctx))
	if err != nil {
		response.NotFound(ctx, "Event not found")
		return
	}
	response.OK(ctx, "Event retrieved successfully", res)
}

// GetEventBySlug godoc
// @Summary Get event by slug
// @Description Public if published; otherwise requires organizer access.
// @Tags events
// @Accept json
// @Produce json
// @Param slug path string true "Event slug"
// @Success 200 {object} responses.EventResponse "Event"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/slug/{slug} [get]
func (c *EventController) GetBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")
	res, err := c.eventService.GetBySlug(ctx.Request.Context(), slug, viewerIDFromCtx(ctx))
	if err != nil {
		response.NotFound(ctx, "Event not found")
		return
	}
	response.OK(ctx, "Event retrieved successfully", res)
}

// UpdateEvent godoc
// @Summary Update event
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.EventUpdateRequest true "Update"
// @Success 200 {object} responses.EventResponse "Event updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id} [put]
func (c *EventController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.EventUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, ok := actorIDFromCtx(ctx)
	if !ok {
		response.Unauthorized(ctx, "Unauthenticated", nil)
		return
	}
	res, err := c.eventService.Update(ctx.Request.Context(), actorID, id, &req)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Forbidden(ctx, "You cannot modify this event")
			return
		}
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Event updated successfully", res)
}

// DeleteEvent godoc
// @Summary Delete event
// @Description Owner only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} responses.Response "Event deleted"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id} [delete]
func (c *EventController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	actorID, ok := actorIDFromCtx(ctx)
	if !ok {
		response.Unauthorized(ctx, "Unauthenticated", nil)
		return
	}
	if err := c.eventService.Delete(ctx.Request.Context(), actorID, id); err != nil {
		if err.Error() == "forbidden" {
			response.Forbidden(ctx, "You cannot delete this event")
			return
		}
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Event deleted successfully", nil)
}

// ----------------------------------------------------------
// Lifecycle
// ----------------------------------------------------------

// PublishEvent godoc
// @Summary Publish event
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} responses.EventResponse "Event published"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/publish [post]
func (c *EventController) Publish(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	actorID, ok := actorIDFromCtx(ctx)
	if !ok {
		response.Unauthorized(ctx, "Unauthenticated", nil)
		return
	}
	res, err := c.eventService.Publish(ctx.Request.Context(), actorID, id)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Event published successfully", res)
}

// UnpublishEvent godoc
// @Summary Unpublish event
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} responses.EventResponse "Event unpublished"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/unpublish [post]
func (c *EventController) Unpublish(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	actorID, ok := actorIDFromCtx(ctx)
	if !ok {
		response.Unauthorized(ctx, "Unauthenticated", nil)
		return
	}
	res, err := c.eventService.Unpublish(ctx.Request.Context(), actorID, id)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Event unpublished successfully", res)
}

// CancelEvent godoc
// @Summary Cancel event
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.EventCancelRequest true "Cancel"
// @Success 200 {object} responses.EventResponse "Event cancelled"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/cancel [post]
func (c *EventController) Cancel(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.EventCancelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, ok := actorIDFromCtx(ctx)
	if !ok {
		response.Unauthorized(ctx, "Unauthenticated", nil)
		return
	}
	res, err := c.eventService.Cancel(ctx.Request.Context(), actorID, id, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Event cancelled successfully", res)
}

// PostponeEvent godoc
// @Summary Postpone event
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.EventPostponeRequest true "Postpone"
// @Success 200 {object} responses.EventResponse "Event postponed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/postpone [post]
func (c *EventController) Postpone(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.EventPostponeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, ok := actorIDFromCtx(ctx)
	if !ok {
		response.Unauthorized(ctx, "Unauthenticated", nil)
		return
	}
	res, err := c.eventService.Postpone(ctx.Request.Context(), actorID, id, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Event postponed successfully", res)
}

// ----------------------------------------------------------
// Occurrences
// ----------------------------------------------------------

// AddOccurrence godoc
// @Summary Add occurrence
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.OccurrenceCreateRequest true "Occurrence"
// @Success 201 {object} responses.EventOccurrenceResponse "Occurrence added"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/occurrences [post]
func (c *EventController) AddOccurrence(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.OccurrenceCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.eventService.AddOccurrence(ctx.Request.Context(), actorID, id, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Occurrence added successfully", res)
}

// ListOccurrences godoc
// @Summary List occurrences
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {array} responses.EventOccurrenceResponse "Occurrences"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/occurrences [get]
func (c *EventController) ListOccurrences(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	res, err := c.eventService.ListOccurrences(ctx.Request.Context(), id)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Occurrences retrieved successfully", res)
}

// UpdateOccurrence godoc
// @Summary Update occurrence
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param occ_id path string true "Occurrence ID"
// @Param request body requests.OccurrenceUpdateRequest true "Update"
// @Success 200 {object} responses.EventOccurrenceResponse "Updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/occurrences/{occ_id} [put]
func (c *EventController) UpdateOccurrence(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	occID, err := uuid.Parse(ctx.Param("occ_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid occurrence id", nil)
		return
	}
	var req requests.OccurrenceUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.eventService.UpdateOccurrence(ctx.Request.Context(), actorID, id, occID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Occurrence updated successfully", res)
}

// DeleteOccurrence godoc
// @Summary Delete occurrence
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param occ_id path string true "Occurrence ID"
// @Success 200 {object} responses.Response "Deleted"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/occurrences/{occ_id} [delete]
func (c *EventController) DeleteOccurrence(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	occID, err := uuid.Parse(ctx.Param("occ_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid occurrence id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	if err := c.eventService.DeleteOccurrence(ctx.Request.Context(), actorID, id, occID); err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Occurrence deleted successfully", nil)
}

// ----------------------------------------------------------
// Co-organizers
// ----------------------------------------------------------

// AddCoOrganizer godoc
// @Summary Add co-organizer
// @Description Owner only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.CoOrganizerAddRequest true "Co-organizer"
// @Success 201 {object} responses.EventCoOrganizerResponse "Added"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/co-organizers [post]
func (c *EventController) AddCoOrganizer(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.CoOrganizerAddRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.eventService.AddCoOrganizer(ctx.Request.Context(), actorID, id, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Co-organizer added successfully", res)
}

// ListCoOrganizers godoc
// @Summary List co-organizers
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {array} responses.EventCoOrganizerResponse "Co-organizers"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/co-organizers [get]
func (c *EventController) ListCoOrganizers(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	res, err := c.eventService.ListCoOrganizers(ctx.Request.Context(), id)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Co-organizers retrieved successfully", res)
}

// RemoveCoOrganizer godoc
// @Summary Remove co-organizer
// @Description Owner only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} responses.Response "Removed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/co-organizers/{user_id} [delete]
func (c *EventController) RemoveCoOrganizer(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid user id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	if err := c.eventService.RemoveCoOrganizer(ctx.Request.Context(), actorID, id, userID); err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Co-organizer removed successfully", nil)
}

// ----------------------------------------------------------
// Staff
// ----------------------------------------------------------

// AddStaff godoc
// @Summary Add staff
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body requests.StaffAddRequest true "Staff"
// @Success 201 {object} responses.EventStaffResponse "Staff added"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/staff [post]
func (c *EventController) AddStaff(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.StaffAddRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	res, err := c.eventService.AddStaff(ctx.Request.Context(), actorID, id, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Staff added successfully", res)
}

// ListStaff godoc
// @Summary List staff
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {array} responses.EventStaffResponse "Staff"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/staff [get]
func (c *EventController) ListStaff(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	res, err := c.eventService.ListStaff(ctx.Request.Context(), id)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Staff retrieved successfully", res)
}

// RemoveStaff godoc
// @Summary Remove staff
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} responses.Response "Removed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/staff/{user_id} [delete]
func (c *EventController) RemoveStaff(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid user id", nil)
		return
	}
	actorID, _ := actorIDFromCtx(ctx)
	if err := c.eventService.RemoveStaff(ctx.Request.Context(), actorID, id, userID); err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Staff removed successfully", nil)
}

// ----------------------------------------------------------
// Follow
// ----------------------------------------------------------

// FollowEvent godoc
// @Summary Follow event
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} responses.Response "Followed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/follow [post]
func (c *EventController) FollowEvent(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	if err := c.eventService.FollowEvent(ctx.Request.Context(), userID, id); err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Event followed successfully", nil)
}

// UnfollowEvent godoc
// @Summary Unfollow event
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} responses.Response "Unfollowed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/follow [delete]
func (c *EventController) UnfollowEvent(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	if err := c.eventService.UnfollowEvent(ctx.Request.Context(), userID, id); err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Event unfollowed successfully", nil)
}

// ListEventFollowers godoc
// @Summary List event followers
// @Description Organizer/co-organizer only.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.EventFollowerResponse "Followers"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/followers [get]
func (c *EventController) ListEventFollowers(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	actorID, _ := actorIDFromCtx(ctx)
	list, total, err := c.eventService.ListEventFollowers(ctx.Request.Context(), actorID, id, page, perPage)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Forbidden(ctx, "You cannot view followers for this event")
			return
		}
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Followers retrieved successfully", list, page, perPage, total)
}

// FollowOrganizer godoc
// @Summary Follow organizer
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organizer_id path string true "Organizer ID"
// @Success 200 {object} responses.Response "Followed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /organizers/{organizer_id}/follow [post]
func (c *EventController) FollowOrganizer(ctx *gin.Context) {
	organizerID, err := uuid.Parse(ctx.Param("organizer_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid organizer id", nil)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	if err := c.eventService.FollowOrganizer(ctx.Request.Context(), userID, organizerID); err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Organizer followed successfully", nil)
}

// UnfollowOrganizer godoc
// @Summary Unfollow organizer
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organizer_id path string true "Organizer ID"
// @Success 200 {object} responses.Response "Unfollowed"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /organizers/{organizer_id}/follow [delete]
func (c *EventController) UnfollowOrganizer(ctx *gin.Context) {
	organizerID, err := uuid.Parse(ctx.Param("organizer_id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid organizer id", nil)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	if err := c.eventService.UnfollowOrganizer(ctx.Request.Context(), userID, organizerID); err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Organizer unfollowed successfully", nil)
}

// ----------------------------------------------------------
// Share
// ----------------------------------------------------------

// RecordShare godoc
// @Summary Record event share
// @Description Public endpoint (optional auth). Records a share for analytics.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param request body requests.ShareRequest true "Channel"
// @Success 201 {object} responses.EventShareResponse "Share recorded"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /events/{id}/share [post]
func (c *EventController) RecordShare(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid event id", nil)
		return
	}
	var req requests.ShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	res, err := c.eventService.RecordShare(ctx.Request.Context(), id, viewerIDFromCtx(ctx), &req, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Created(ctx, "Share recorded successfully", res)
}
