package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"sevensolutions-backend/internal/application"
	"sevensolutions-backend/internal/domain"
	"sevensolutions-backend/pkg/response"
)

type AuthHandler struct {
	authService *application.AuthService
}

func NewAuthHandler(authService *application.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	input := domain.RegisterInput{Name: req.Name, Email: req.Email, Password: req.Password}

	user, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrEmailExists) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, toUserResponse(user))
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "email and password are required")
		return
	}

	input := domain.LoginInput{Email: req.Email, Password: req.Password}

	token, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"token": token})
}

// แยก 400 (ขาด field) กับ 422 (format ผิด) จาก validator tag เดียวกัน
func bindError(c *gin.Context, err error) {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		response.BadRequest(c, "invalid request body")
		return
	}

	for _, fieldErr := range validationErrs {
		if fieldErr.Tag() == "required" {
			response.BadRequest(c, "missing required fields")
			return
		}
	}

	response.ValidationError(c, "email format invalid or password too short")
}
