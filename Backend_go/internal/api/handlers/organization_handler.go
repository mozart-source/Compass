package handlers

import (
	"net/http"
	"strconv"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/organization"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrganizationHandler handles HTTP requests for organization operations
type OrganizationHandler struct {
	service organization.Service
}

// NewOrganizationHandler creates a new OrganizationHandler instance
func NewOrganizationHandler(service organization.Service) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

// CreateOrganization godoc
// @Summary Create a new organization
// @Description Create a new organization with the provided information
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization body dto.CreateOrganizationRequest true "Organization creation request"
// @Success 201 {object} dto.OrganizationResponse "Organization created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 409 {object} map[string]string "Organization name already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req dto.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Convert userID to UUID - it's already a UUID, no need to parse
	creatorID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
		return
	}

	input := organization.CreateOrganizationInput{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		CreatorID:   creatorID,
		OwnerID:     creatorID, // Initially, creator is also the owner
	}

	createdOrg, err := h.service.CreateOrganization(c.Request.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == organization.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		} else if err == organization.ErrDuplicateName {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dto.OrganizationToResponse(createdOrg)})
}

// GetOrganization godoc
// @Summary Get an organization by ID
// @Description Get detailed information about a specific organization
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID" format(uuid)
// @Success 200 {object} dto.OrganizationResponse "Organization details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid organization ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Organization not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	org, err := h.service.GetOrganization(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == organization.ErrOrganizationNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.OrganizationToResponse(org)})
}

// GetOrganizationStats godoc
// @Summary Get organization statistics
// @Description Get detailed statistics about a specific organization
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID" format(uuid)
// @Success 200 {object} dto.OrganizationStatsResponse "Organization statistics retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid organization ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Organization not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/organizations/{id}/stats [get]
func (h *OrganizationHandler) GetOrganizationStats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	org, err := h.service.GetOrganization(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == organization.ErrOrganizationNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement getting actual stats from service
	stats := dto.OrganizationStatsResponse{
		Organization:  *dto.OrganizationToResponse(org),
		MembersCount:  0, // To be implemented
		ProjectsCount: 0, // To be implemented
		TasksCount:    0, // To be implemented
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// ListOrganizations godoc
// @Summary List all organizations
// @Description Get a paginated list of organizations
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 0)"
// @Param pageSize query int false "Number of items per page (default: 10)"
// @Success 200 {object} dto.OrganizationListResponse "List of organizations retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid pagination parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/organizations [get]
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
		return
	}

	filter := organization.OrganizationFilter{
		Page:     page,
		PageSize: pageSize,
	}

	organizations, total, err := h.service.ListOrganizations(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]dto.OrganizationResponse, len(organizations))
	for i, org := range organizations {
		response := dto.OrganizationToResponse(&org)
		responses[i] = *response
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.OrganizationListResponse{
		Organizations: responses,
		TotalCount:    total,
		Page:          page,
		PageSize:      pageSize,
	}})
}

// UpdateOrganization godoc
// @Summary Update an organization
// @Description Update an existing organization's information
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID" format(uuid)
// @Param organization body dto.UpdateOrganizationRequest true "Organization update information"
// @Success 200 {object} dto.OrganizationResponse "Organization updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or organization ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Organization not found"
// @Failure 409 {object} map[string]string "Organization name already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	var req dto.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Get the organization to check ownership
	org, err := h.service.GetOrganization(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == organization.ErrOrganizationNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Only the owner can update the organization
	if org.OwnerID.String() != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the organization owner can update it"})
		return
	}

	input := organization.UpdateOrganizationInput{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		OwnerID:     req.OwnerID,
	}

	updatedOrg, err := h.service.UpdateOrganization(c.Request.Context(), id, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == organization.ErrOrganizationNotFound {
			statusCode = http.StatusNotFound
		} else if err == organization.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		} else if err == organization.ErrDuplicateName {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.OrganizationToResponse(updatedOrg)})
}

// DeleteOrganization godoc
// @Summary Delete an organization
// @Description Delete an existing organization
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID" format(uuid)
// @Success 204 "Organization deleted successfully"
// @Failure 400 {object} map[string]string "Invalid organization ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Not the organization owner"
// @Failure 404 {object} map[string]string "Organization not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Get the organization to check ownership
	org, err := h.service.GetOrganization(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == organization.ErrOrganizationNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Only the owner can delete the organization
	if org.OwnerID.String() != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the organization owner can delete it"})
		return
	}

	err = h.service.DeleteOrganization(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == organization.ErrOrganizationNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
