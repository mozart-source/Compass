package handlers

import (
	"net/http"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/roles"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles HTTP requests for auth operations
type AuthHandler struct {
	service roles.Service
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(service roles.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with the provided information
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role body dto.CreateRoleRequest true "Role creation request"
// @Success 201 {object} dto.RoleResponse "Role created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/roles [post]
func (h *AuthHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := roles.CreateRoleInput{
		Name:        req.Name,
		Description: req.Description,
	}

	role, err := h.service.CreateRole(c.Request.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == roles.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dto.RoleToResponse(role)})
}

// GetRole godoc
// @Summary Get a role by ID
// @Description Get detailed information about a specific role
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} dto.RoleResponse "Role details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid role ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Role not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/roles/{id} [get]
func (h *AuthHandler) GetRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	role, err := h.service.GetRole(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == roles.ErrRoleNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.RoleToResponse(role)})
}

// ListRoles godoc
// @Summary List all roles
// @Description Get a list of all roles
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.RoleResponse "List of roles retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/roles [get]
func (h *AuthHandler) ListRoles(c *gin.Context) {
	roles, err := h.service.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = *dto.RoleToResponse(&role)
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateRole godoc
// @Summary Update a role
// @Description Update an existing role's information
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID" format(uuid)
// @Param role body dto.UpdateRoleRequest true "Role update information"
// @Success 200 {object} dto.RoleResponse "Role updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or role ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Role not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/roles/{id} [put]
func (h *AuthHandler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := roles.UpdateRoleInput{
		Name:        req.Name,
		Description: req.Description,
	}

	role, err := h.service.UpdateRole(c.Request.Context(), id, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == roles.ErrRoleNotFound {
			statusCode = http.StatusNotFound
		} else if err == roles.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.RoleToResponse(role)})
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Delete an existing role
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID" format(uuid)
// @Success 204 "Role deleted successfully"
// @Failure 400 {object} map[string]string "Invalid role ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Role not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/roles/{id} [delete]
func (h *AuthHandler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	if err := h.service.DeleteRole(c.Request.Context(), id); err != nil {
		statusCode := http.StatusInternalServerError
		if err == roles.ErrRoleNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AssignPermissionToRole godoc
// @Summary Assign a permission to a role
// @Description Assign a permission to an existing role
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID" format(uuid)
// @Param permission_id path string true "Permission ID" format(uuid)
// @Success 204 "Permission assigned successfully"
// @Failure 400 {object} map[string]string "Invalid role or permission ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Role or permission not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/roles/{id}/permissions/{permission_id} [post]
func (h *AuthHandler) AssignPermissionToRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	permissionID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission ID"})
		return
	}

	if err := h.service.AssignPermissionToRole(c.Request.Context(), roleID, permissionID); err != nil {
		statusCode := http.StatusInternalServerError
		if err == roles.ErrRoleNotFound || err == roles.ErrPermissionNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RemovePermissionFromRole godoc
// @Summary Remove a permission from a role
// @Description Remove a permission from an existing role
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID" format(uuid)
// @Param permission_id path string true "Permission ID" format(uuid)
// @Success 204 "Permission removed successfully"
// @Failure 400 {object} map[string]string "Invalid role or permission ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Role or permission not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/roles/{id}/permissions/{permission_id} [delete]
func (h *AuthHandler) RemovePermissionFromRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	permissionID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission ID"})
		return
	}

	if err := h.service.RemovePermissionFromRole(c.Request.Context(), roleID, permissionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AssignRoleToUser godoc
// @Summary Assign a role to a user
// @Description Assign a role to an existing user
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID" format(uuid)
// @Param role_id path string true "Role ID" format(uuid)
// @Success 204 "Role assigned successfully"
// @Failure 400 {object} map[string]string "Invalid user or role ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "User or role not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/users/{user_id}/roles/{role_id} [post]
func (h *AuthHandler) AssignRoleToUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	if err := h.service.AssignRoleToUser(c.Request.Context(), userID, roleID); err != nil {
		statusCode := http.StatusInternalServerError
		if err == roles.ErrRoleNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUserRoles godoc
// @Summary Get user roles
// @Description Get all roles assigned to a user
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID" format(uuid)
// @Success 200 {array} dto.RoleResponse "List of user roles retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/users/{user_id}/roles [get]
func (h *AuthHandler) GetUserRoles(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	roles, err := h.service.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = *dto.RoleToResponse(&role)
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}
