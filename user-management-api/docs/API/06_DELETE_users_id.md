# DELETE /api/users/:id - Delete user

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
2. Parse `:id` -> `400` if invalid ObjectID
3. DeleteOne by `_id` -> `404` if DeletedCount == 0
4. Return success message

---

## Response `200`

```json
{
  "success": true,
  "data": {
    "message": "user deleted successfully"
  }
}
```

---

## Errors

| Code | Condition |
|------|-----------|
| `400` | Invalid ObjectID format |
| `401` | Missing or invalid token |
| `404` | User not found |
| `500` | MongoDB delete failed |
