package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sotremont/taskflow/internal/domain"
	"github.com/sotremont/taskflow/internal/dto"
	"github.com/sotremont/taskflow/internal/service"
)

type TaskHandler struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.taskService.CreateTask(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *TaskHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	filters := make(map[string]interface{})
	if teamID := c.Query("team_id"); teamID != "" {
		if id, err := uuid.Parse(teamID); err == nil {
			filters["team_id"] = id
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = domain.TaskStatus(status)
	}
	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = domain.TaskPriority(priority)
	}
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		if id, err := uuid.Parse(assigneeID); err == nil {
			filters["assignee_id"] = id
		}
	}

	resp, err := h.taskService.ListTasks(c.Request.Context(), userID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) Get(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	resp, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.taskService.UpdateTask(c.Request.Context(), userID, taskID, req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	if err := h.taskService.DeleteTask(c.Request.Context(), userID, taskID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *TaskHandler) Assign(c *gin.Context) {
	operatorID := c.MustGet("user_id").(uuid.UUID)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var req dto.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.AssignTask(c.Request.Context(), operatorID, taskID, req.AssigneeID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task assigned"})
}

func (h *TaskHandler) GetAnalytics(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	resp, err := h.taskService.GetAnalytics(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) GetHistory(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	resp, err := h.taskService.GetHistory(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
