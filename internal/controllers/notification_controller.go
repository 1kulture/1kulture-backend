package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/response"

	_ "github.com/1kulture/1kulture-backend/internal/responses"
)

type NotificationController struct {
	service serviceInterfaces.NotificationService
}

func NewNotificationController(s serviceInterfaces.NotificationService) *NotificationController {
	return &NotificationController{service: s}
}

// ListNotifications godoc
// @Summary List my notifications
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param unread_only query bool false "Only unread"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.NotificationResponse "Notifications"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /notifications [get]
func (c *NotificationController) List(ctx *gin.Context) {
	unreadOnly := ctx.DefaultQuery("unread_only", "false") == "true"
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	userID, _ := actorIDFromCtx(ctx)
	list, total, err := c.service.List(ctx.Request.Context(), userID, unreadOnly, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Notifications retrieved", list, page, perPage, total)
}

// CountUnread godoc
// @Summary Count unread notifications
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.Response "Count"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /notifications/unread-count [get]
func (c *NotificationController) CountUnread(ctx *gin.Context) {
	userID, _ := actorIDFromCtx(ctx)
	count, err := c.service.CountUnread(ctx.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "Unread count retrieved", gin.H{"count": count})
}

// MarkNotificationRead godoc
// @Summary Mark a notification as read
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} responses.Response "Marked"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /notifications/{id}/read [put]
func (c *NotificationController) MarkRead(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid notification id", nil)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	if err := c.service.MarkRead(ctx.Request.Context(), userID, id); err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Notification marked as read", nil)
}

// MarkAllNotificationsRead godoc
// @Summary Mark all notifications as read
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.Response "Marked"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /notifications/read-all [put]
func (c *NotificationController) MarkAllRead(ctx *gin.Context) {
	userID, _ := actorIDFromCtx(ctx)
	if err := c.service.MarkAllRead(ctx.Request.Context(), userID); err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.OK(ctx, "All notifications marked as read", nil)
}

// DeleteNotification godoc
// @Summary Delete a notification
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} responses.Response "Deleted"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /notifications/{id} [delete]
func (c *NotificationController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid notification id", nil)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	if err := c.service.Delete(ctx.Request.Context(), userID, id); err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Notification deleted", nil)
}
