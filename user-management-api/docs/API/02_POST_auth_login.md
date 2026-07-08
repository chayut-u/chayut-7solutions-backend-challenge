# POST /api/auth/login - Login and get JWT

**← [_ROUTES.md](./_ROUTES.md)** | Spec: [FUNC §3](../FUNC_UserManagement.md#3-authentication-rules)

---

## Permission

Public endpoint - no JWT required.

---

## Request Body

```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

| Field | Type | Required |
|-------|------|:--------:|
| `email` | string | ✅ |
| `password` | string | ✅ |

---

## Logic

1. Bind + validate request body -> `400` if missing fields
2. Find user by email -> `401` if not found (do not reveal "email not found")
3. Compare bcrypt hash -> `401` if mismatch
4. Sign JWT: `sub` = userID (hex), `exp` = now + 24h, algorithm = HS256
5. Return token

---

## Response `200`

```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

---

## Errors

| Code | Condition |
|------|-----------|
| `400` | Missing email or password |
| `401` | Email not found OR password mismatch (same message - do not leak which) |
| `500` | Unexpected server error |

---

## Note on Security

Return the same `401` message for both "email not found" and "wrong password" to prevent user enumeration:
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid credentials"
  }
}
```
