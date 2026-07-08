# GET /api/users/:id - Get user by ID

**← [_ROUTES.md](./_ROUTES.md)** | Spec: [FUNC §4](../FUNC_UserManagement.md#4-business-rules)

---

## Permission

Requires valid JWT.

---

## Path Params

| Param | Type | Description |
|-------|------|-------------|
| `id` | string | MongoDB ObjectID hex string |

---

## Logic

1. Validate JWT -> `401` if invalid
2. Parse `:id` as `primitive.ObjectID` -> `400` if not valid ObjectID format
3. Find document by `_id` -> `404` if not found
4. Return user (exclude password)

---

## Response `200`

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
| `400` | `:id` is not a valid ObjectID |
| `401` | Missing or invalid token |
| `404` | User not found |
| `500` | MongoDB query failed |
