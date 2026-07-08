package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"sevensolutions-backend/internal/application"
	"sevensolutions-backend/internal/domain"
	"sevensolutions-backend/pkg/response"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

type UserHandler struct {
	userService *application.UserService
}

func NewUserHandler(userService *application.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.userService.GetAll(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	items := make([]UserResponse, len(result.Items))
	for i, user := range result.Items {
		items[i] = toUserResponse(user)
	}

	response.OK(c, gin.H{
		"items": items,
		"total": result.Total,
		"page":  result.Page,
		"limit": result.Limit,
	})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	user, err := h.userService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		mapUserError(c, err)
		return
	}

	response.OK(c, toUserResponse(user))
}

type updateUserRequest struct {
	Name  *string `json:"name" binding:"omitempty,min=1"`
	Email *string `json:"email" binding:"omitempty,email"`
}

func (h *UserHandler) Update(c *gin.Context) {
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	input := domain.UpdateUserInput{Name: req.Name, Email: req.Email}

	user, err := h.userService.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		mapUserError(c, err)
		return
	}

	response.OK(c, toUserResponse(user))
}

func (h *UserHandler) Delete(c *gin.Context) {
	err := h.userService.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		mapUserError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "user deleted successfully"})
}

// map sentinel error เป็น HTTP status ใช้ร่วมกันทุก handler ที่มี user ID
func mapUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, application.ErrInvalidUserID):
		response.BadRequest(c, err.Error())
	case errors.Is(err, application.ErrNoFieldsToUpdate):
		response.BadRequest(c, err.Error())
	case errors.Is(err, application.ErrUserNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, application.ErrEmailExists):
		response.Conflict(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
