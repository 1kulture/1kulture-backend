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

type BrandController struct {
	service serviceInterfaces.BrandProfileService
}

func NewBrandController(s serviceInterfaces.BrandProfileService) *BrandController {
	return &BrandController{service: s}
}

// CreateBrandProfile godoc
// @Summary Create brand profile
// @Description Create the brand profile for the authenticated user. Assigns the "brand" role.
// @Tags brands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateBrandProfileRequest true "Brand profile"
// @Success 201 {object} responses.BrandProfileResponse "Brand profile created"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 409 {object} responses.ErrorResponse "Already exists"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands/profile [post]
func (c *BrandController) Create(ctx *gin.Context) {
	var req requests.CreateBrandProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.CreateForUser(ctx.Request.Context(), userID, &req)
	if err != nil {
		logger.WithRequest(ctx).Error("CreateBrandProfile: ", err)
		response.Conflict(ctx, err.Error(), nil)
		return
	}
	response.Created(ctx, "Brand profile created successfully", res)
}

// GetMyBrandProfile godoc
// @Summary Get my brand profile
// @Tags brands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.BrandProfileResponse "Brand profile"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands/profile [get]
func (c *BrandController) GetMine(ctx *gin.Context) {
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.GetMine(ctx.Request.Context(), userID)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Brand profile retrieved", res)
}

// UpdateMyBrandProfile godoc
// @Summary Update my brand profile
// @Tags brands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.UpdateBrandProfileRequest true "Update"
// @Success 200 {object} responses.BrandProfileResponse "Updated"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 422 {object} responses.ErrorResponse "Validation error"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands/profile [put]
func (c *BrandController) UpdateMine(ctx *gin.Context) {
	var req requests.UpdateBrandProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", err.Error())
		return
	}
	if errs := validator.Struct(req); errs != nil {
		response.ValidationError(ctx, errs)
		return
	}
	userID, _ := actorIDFromCtx(ctx)
	res, err := c.service.Update(ctx.Request.Context(), userID, &req)
	if err != nil {
		response.BadRequest(ctx, err.Error(), nil)
		return
	}
	response.OK(ctx, "Brand profile updated", res)
}

// GetBrandProfileByID godoc
// @Summary Get a brand profile by ID
// @Description Organizers can view brand profiles when evaluating partnership requests.
// @Tags brands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Brand Profile ID"
// @Success 200 {object} responses.BrandProfileResponse "Brand profile"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 404 {object} responses.ErrorResponse "Not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands/{id} [get]
func (c *BrandController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "Invalid brand id", nil)
		return
	}
	res, err := c.service.GetByID(ctx.Request.Context(), id)
	if err != nil {
		response.NotFound(ctx, err.Error())
		return
	}
	response.OK(ctx, "Brand profile retrieved", res)
}

// ListBrandProfiles godoc
// @Summary List brand profiles
// @Description Browse brands. Filters: industry, location.
// @Tags brands
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param industry query string false "Industry"
// @Param location query string false "Location"
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Success 200 {array} responses.BrandProfileResponse "Brands"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /brands [get]
func (c *BrandController) List(ctx *gin.Context) {
	industry := ctx.Query("industry")
	location := ctx.Query("location")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))
	list, total, err := c.service.List(ctx.Request.Context(), industry, location, page, perPage)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Paginated(ctx, "Brands retrieved", list, page, perPage, total)
}
