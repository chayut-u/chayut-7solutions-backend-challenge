# PUT /api/users/:id - Update user

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

## Request Body

At least one field required.

```json
{
  "name": "John Updated",
  "email": "john.updated@example.com"
}
```

| Field | Type | Required | Validation |
|-------|------|:--------:|------------|
| `name` | string | - | non-empty if provided |
| `email` | string | - | valid email format if provided |

---

## Logic

1. Validate JWT -> `401` if invalid
2. Parse `:id` -> `400` if invalid ObjectID
3. Bind body -> `400` if both name and email are missing
4. If email provided: check uniqueness against other users -> `409` if taken
5. Build MongoDB `$set` update document with only provided fields
6. FindOneAndUpdate with `ReturnDocument = After` -> `404` if not found
7. Return updated user

---

## Response `200`

```json
{
  "success": true,
  "data": {
    "id": "64b1f2c3d4e5f6a7b8c9d0e1",
    "name": "John Updated",
    "email": "john.updated@example.com",
    "created_at": "2026-07-07T10:00:00Z"
  }
}
```

---

## Errors

| Code | Condition |
|------|-----------|
| `400` | Invalid ObjectID or no fields to update |
| `401` | Missing or invalid token |
| `404` | User not found |
| `409` | New email already in use by another user |
| `500` | MongoDB update failed |
