package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/todos"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TodoHandler struct {
	service todos.Service
}

func NewTodoHandler(service todos.Service) *TodoHandler {
	return &TodoHandler{service: service}
}

// CreateTodo godoc
// @Summary Create a new todo
// @Description Create a new todo with the provided information
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param todo body dto.CreateTodoRequest true "Todo creation request"
// @Success 201 {object} dto.TodoResponse "Todo created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos [post]
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req dto.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Convert and validate status
	status := todos.TodoStatus(req.Status)
	if !status.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	// Convert and validate priority
	priority := todos.TodoPriority(req.Priority)
	if !priority.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority value"})
		return
	}

	// Get or create default list if no list ID provided
	var listID uuid.UUID
	if req.ListID == nil {
		defaultList, err := h.service.GetOrCreateDefaultList(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get default list"})
			return
		}
		listID = defaultList.ID
	} else {
		listID = *req.ListID
	}

	input := todos.CreateTodoInput{
		Title:                 req.Title,
		Description:           req.Description,
		Status:                status,
		Priority:              priority,
		DueDate:               req.DueDate,
		ReminderTime:          req.ReminderTime,
		IsRecurring:           req.IsRecurring,
		RecurrencePattern:     req.RecurrencePattern,
		Tags:                  req.Tags,
		Checklist:             req.Checklist,
		LinkedTaskID:          req.LinkedTaskID,
		LinkedCalendarEventID: req.LinkedCalendarEventID,
		UserID:                userID,
		ListID:                listID,
		IsCompleted:           req.IsCompleted,
	}

	createdTodo, err := h.service.CreateTodo(c.Request.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": TodoToResponse(createdTodo)})
}

// GetTodo godoc
// @Summary Get a todo by ID
// @Description Get detailed information about a specific todo
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID" format(uuid)
// @Success 200 {object} dto.TodoResponse "Todo details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid todo ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos/{id} [get]
func (h *TodoHandler) GetTodo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	todo, err := h.service.GetTodo(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TodoToResponse(todo)})
}

// ListTodos godoc
// @Summary List all todos
// @Description Get a paginated list of todos with optional filters
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 0)"
// @Param pageSize query int false "Number of items per page (default: 10)"
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param is_completed query bool false "Filter by completion status"
// @Success 200 {object} dto.TodoListResponse "List of todos retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos [get]
func (h *TodoHandler) ListTodos(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
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

	filter := todos.TodoFilter{
		Page:     page,
		PageSize: pageSize,
		UserID:   &userID,
	}

	// Parse optional filters
	if statusStr := c.Query("status"); statusStr != "" {
		status := todos.TodoStatus(statusStr)
		if status.IsValid() {
			filter.Status = &status
		}
	}
	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority := todos.TodoPriority(priorityStr)
		if priority.IsValid() {
			filter.Priority = &priority
		}
	}
	if isCompletedStr := c.Query("is_completed"); isCompletedStr != "" {
		isCompleted := isCompletedStr == "true"
		filter.IsCompleted = &isCompleted
	}

	todosList, total, err := h.service.ListTodos(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.TodoListResponse{
		Todos:      TodosToResponse(todosList),
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateTodo godoc
// @Summary Update a todo
// @Description Update an existing todo's information
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID" format(uuid)
// @Param todo body dto.UpdateTodoRequest true "Todo update information"
// @Success 200 {object} dto.TodoResponse "Todo updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos/{id} [put]
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	var req dto.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := todos.UpdateTodoInput{
		Title:                 req.Title,
		Description:           req.Description,
		DueDate:               req.DueDate,
		ReminderTime:          req.ReminderTime,
		IsRecurring:           req.IsRecurring,
		RecurrencePattern:     map[string]interface{}{},
		Tags:                  map[string]interface{}{},
		Checklist:             map[string]interface{}{},
		LinkedTaskID:          req.LinkedTaskID,
		LinkedCalendarEventID: req.LinkedCalendarEventID,
	}

	if req.RecurrencePattern != nil {
		input.RecurrencePattern = *req.RecurrencePattern
	}
	if req.Tags != nil {
		input.Tags = *req.Tags
	}
	if req.Checklist != nil {
		input.Checklist = *req.Checklist
	}

	// Convert status if provided
	if req.Status != nil {
		status := todos.TodoStatus(*req.Status)
		if !status.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
			return
		}
		input.Status = &status
	}

	// Convert priority if provided
	if req.Priority != nil {
		priority := todos.TodoPriority(*req.Priority)
		if !priority.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority value"})
			return
		}
		input.Priority = &priority
	}

	updatedTodo, err := h.service.UpdateTodo(c.Request.Context(), id, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		} else if err == todos.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TodoToResponse(updatedTodo)})
}

// DeleteTodo godoc
// @Summary Delete a todo
// @Description Delete an existing todo
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID" format(uuid)
// @Success 204 "Todo deleted successfully"
// @Failure 400 {object} map[string]string "Invalid todo ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos/{id} [delete]
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	err = h.service.DeleteTodo(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateTodoStatus godoc
// @Summary Update todo status
// @Description Update the status of a todo
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID" format(uuid)
// @Param status body dto.UpdateTodoStatusRequest true "Todo status update information"
// @Success 200 {object} dto.TodoResponse "Todo status updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos/{id}/status [patch]
func (h *TodoHandler) UpdateTodoStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	var req dto.UpdateTodoStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := todos.TodoStatus(req.Status)
	if !status.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	updatedTodo, err := h.service.UpdateTodoStatus(c.Request.Context(), id, status)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		} else if err == todos.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TodoToResponse(updatedTodo)})
}

// UpdateTodoPriority godoc
// @Summary Update todo priority
// @Description Update the priority of a todo
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID" format(uuid)
// @Param priority body dto.UpdateTodoPriorityRequest true "Todo priority update information"
// @Success 200 {object} dto.TodoResponse "Todo priority updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos/{id}/priority [patch]
func (h *TodoHandler) UpdateTodoPriority(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	var req dto.UpdateTodoPriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priority := todos.TodoPriority(req.Priority)
	if !priority.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority value"})
		return
	}

	updatedTodo, err := h.service.UpdateTodoPriority(c.Request.Context(), id, priority)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		} else if err == todos.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TodoToResponse(updatedTodo)})
}

// CompleteTodo godoc
// @Summary Complete a todo
// @Description Mark a todo as completed
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID" format(uuid)
// @Success 200 {object} dto.TodoResponse "Todo marked as completed successfully"
// @Failure 400 {object} map[string]string "Invalid todo ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos/{id}/complete [patch]
func (h *TodoHandler) CompleteTodo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	updatedTodo, err := h.service.CompleteTodo(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TodoToResponse(updatedTodo)})
}

// UncompleteTodo godoc
// @Summary Uncomplete a todo
// @Description Mark a todo as uncompleted
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo ID" format(uuid)
// @Success 200 {object} dto.TodoResponse "Todo marked as uncompleted successfully"
// @Failure 400 {object} map[string]string "Invalid todo ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos/{id}/uncomplete [patch]
func (h *TodoHandler) UncompleteTodo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	updatedTodo, err := h.service.UncompleteTodo(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TodoToResponse(updatedTodo)})
}

// CreateTodoList godoc
// @Summary Create a new todo list
// @Description Create a new todo list with the provided information
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param todoList body todos.CreateTodoListInput true "Todo list creation request"
// @Success 201 {object} dto.TodoListResponse "Todo list created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todo-lists [post]
func (h *TodoHandler) CreateTodoList(c *gin.Context) {
	var input todos.CreateTodoListInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	input.UserID = userID

	todoList := &todos.TodoList{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		UserID:      input.UserID,
		IsDefault:   input.IsDefault,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.service.CreateTodoList(c.Request.Context(), todoList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": TodoListToResponse(todoList)})
}

// GetTodosByUser godoc
// @Summary Get todos by user ID
// @Description Get all todos for a specific user
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID" format(uuid)
// @Success 200 {object} dto.UserTodosResponse "Todos retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todos/user/{user_id} [get]
func (h *TodoHandler) GetTodosByUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get current user ID from context (set by auth middleware)
	currentUserID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Only allow users to view their own todos
	if currentUserID != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized to view other user's todos"})
		return
	}

	todos, err := h.service.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.UserTodosResponse{
		Todos:      TodosToResponse(todos),
		TotalCount: int64(len(todos)),
		Page:       1,
		PageSize:   len(todos),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetTodoList godoc
// @Summary Get a todo list by ID
// @Description Get detailed information about a specific todo list
// @Tags todo-lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo List ID" format(uuid)
// @Success 200 {object} dto.TodoListResponse "Todo list details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid todo list ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo list not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todo-lists/{id} [get]
func (h *TodoHandler) GetTodoList(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo list ID"})
		return
	}

	list, err := h.service.GetTodoList(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TodoListToResponse(list)})
}

// GetAllTodoLists godoc
// @Summary Get all todo lists for a user
// @Description Get all todo lists for the authenticated user
// @Tags todo-lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.TodoListsResponse "Todo lists retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todo-lists [get]
func (h *TodoHandler) GetAllTodoLists(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	lists, err := h.service.GetAllTodoLists(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.TodoListsResponse{
		Lists: make([]dto.TodoListResponse, len(lists)),
	}
	for i, list := range lists {
		response.Lists[i] = *TodoListToResponse(&list)
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateTodoList godoc
// @Summary Update a todo list
// @Description Update an existing todo list's information
// @Tags todo-lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo List ID" format(uuid)
// @Param todoList body todos.UpdateTodoListInput true "Todo list update information"
// @Success 200 {object} dto.TodoListResponse "Todo list updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo list not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todo-lists/{id} [put]
func (h *TodoHandler) UpdateTodoList(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo list ID"})
		return
	}

	var input todos.UpdateTodoListInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list, err := h.service.UpdateTodoList(c.Request.Context(), id, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TodoListToResponse(list)})
}

// DeleteTodoList godoc
// @Summary Delete a todo list
// @Description Delete an existing todo list
// @Tags todo-lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Todo List ID" format(uuid)
// @Success 204 "Todo list deleted successfully"
// @Failure 400 {object} map[string]string "Invalid todo list ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Todo list not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/todo-lists/{id} [delete]
func (h *TodoHandler) DeleteTodoList(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo list ID"})
		return
	}

	err = h.service.DeleteTodoList(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == todos.ErrTodoNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
