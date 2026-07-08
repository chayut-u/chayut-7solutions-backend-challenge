# POST /api/auth/register - Register new user

**← [_ROUTES.md](./_ROUTES.md)** | Spec: [FUNC §3](../FUNC_UserManagement.md#3-authentication-rules)

---

## Permission

Public endpoint - no JWT required.

---

## Request Body

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "secret123"
}
```

| Field | Type | Required | Validation |
|-------|------|:--------:|------------|
| `name` | string | ✅ | non-empty |
| `email` | string | ✅ | valid email format |
| `password` | string | ✅ | min 6 characters |

---

## Logic

1. Bind + validate request body -> `400` if invalid
2. Check email uniqueness -> `409` if exists
3. Hash password with bcrypt (cost=10)
4. Save User document to MongoDB
5. Return created user (without password)

---

## Response `201`

```json
{
  "success": true,
  "data": {
    "id": "64b1f2c3d4e5f6a7b8c9d0e1",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-07-07T10:00:00Z"
  }
}
```

---

## Errors

| Code | Condition |
|------|-----------|
| `400` | Missing name, email, or password |
| `409` | Email already registered |
| `422` | Email format invalid or password < 6 chars |
| `500` | MongoDB write failed |
