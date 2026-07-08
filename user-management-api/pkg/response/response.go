package response

import "github.com/gin-gonic/gin"

type envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *errorBody `json:"error,omitempty"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data any) {
	c.JSON(200, envelope{Success: true, Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(201, envelope{Success: true, Data: data})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(400, envelope{Success: false, Error: &errorBody{Code: "BAD_REQUEST", Message: message}})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(401, envelope{Success: false, Error: &errorBody{Code: "UNAUTHORIZED", Message: message}})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(404, envelope{Success: false, Error: &errorBody{Code: "NOT_FOUND", Message: message}})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(409, envelope{Success: false, Error: &errorBody{Code: "CONFLICT", Message: message}})
}

func ValidationError(c *gin.Context, message string) {
	c.JSON(422, envelope{Success: false, Error: &errorBody{Code: "VALIDATION_ERROR", Message: message}})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(500, envelope{Success: false, Error: &errorBody{Code: "INTERNAL_ERROR", Message: message}})
}
