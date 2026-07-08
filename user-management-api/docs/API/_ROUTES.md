# API Routes - User Management

**Spec:** [FUNC_UserManagement.md](../FUNC_UserManagement.md)  
**Stack:** Go (Gin) · JWT (HS256) · MongoDB

---

## Route Table

| # | Method | Path | File | Auth | Status |
|---|--------|------|------|------|--------|
| 01 | `POST` | `/api/auth/register` | [01_POST_auth_register.md](./01_POST_auth_register.md) | ❌ | ⬜ |
| 02 | `POST` | `/api/auth/login` | [02_POST_auth_login.md](./02_POST_auth_login.md) | ❌ | ⬜ |
| 03 | `GET` | `/api/users` | [03_GET_users.md](./03_GET_users.md) | ✅ | ⬜ |
| 04 | `GET` | `/api/users/:id` | [04_GET_users_id.md](./04_GET_users_id.md) | ✅ | ⬜ |
| 05 | `PUT` | `/api/users/:id` | [05_PUT_users_id.md](./05_PUT_users_id.md) | ✅ | ⬜ |
| 06 | `DELETE` | `/api/users/:id` | [06_DELETE_users_id.md](./06_DELETE_users_id.md) | ✅ | ⬜ |
| 07 | `POST` | `/api/users` | [07_POST_users.md](./07_POST_users.md) | ✅ | ⬜ |

> Note: "User registration" (Authentication) และ "Create a new user" (User Operations) ใน spec ทำ logic เดียวกัน (`AuthHandler.Register`) แต่เปิด 2 route: `POST /api/auth/register` (public, self-service) และ `POST /api/users` (ต้อง JWT, ตรงกับ User Operations "Create a new user" ที่อยู่ในกลุ่ม CRUD ที่ protect ด้วย token)

---

## Shared Context

### Auth Middleware

Applied to all `/api/users/*` routes via Gin group:

```go
authorized := r.Group("/api/users")
authorized.Use(middleware.AuthMiddleware(jwtService))
```

Header format: `Authorization: Bearer <token>`

Middleware behavior:
1. Extract `Authorization` header
2. Parse and validate JWT (HS256, check expiry)
3. Set `userID` in Gin context: `c.Set("userID", claims.Subject)`
4. On failure -> 401 UNAUTHORIZED, abort

### Standard Response Helpers

```go
// pkg/response/response.go
func OK(c *gin.Context, data any)
func Created(c *gin.Context, data any)
func BadRequest(c *gin.Context, message string)
func Unauthorized(c *gin.Context, message string)
func NotFound(c *gin.Context, message string)
func Conflict(c *gin.Context, message string)
func InternalError(c *gin.Context, message string)
```

### User Response Shape (never include password)

```go
type UserResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

---

## Status Legend
⬜ ยังไม่เริ่ม | 🔵 กำลังทำ | ✅ เสร็จแล้ว | 🔴 มีปัญหา
