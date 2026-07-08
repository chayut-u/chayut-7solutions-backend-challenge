# POST /api/users - Create user (authenticated)

**← [_ROUTES.md](./_ROUTES.md)** | Spec: [FUNC §4](../FUNC_UserManagement.md#4-business-rules)

---

## Permission

Requires valid JWT (`Authorization: Bearer <token>`).

---

## Logic

ใช้ handler เดียวกับ `POST /api/auth/register` (`AuthHandler.Register`) ทุกอย่างเหมือนกันทั้ง validation,
uniqueness check, และ response shape - ต่างกันแค่ route นี้อยู่หลัง auth middleware

เหตุผลที่มี route แยก: functional spec แบ่ง "User registration" (หมวด Authentication, public)
กับ "Create a new user" (หมวด User Operations ที่อยู่รวมกับ CRUD อื่น ๆ ซึ่ง protect ด้วย JWT ทั้งหมด)
route นี้จึงเปิดให้ operation เดียวกันใช้งานได้ทั้งสองบริบท โดยไม่ซ้ำ business logic

---

## Request Body

```json
{ "name": "Jane Doe", "email": "jane@example.com", "password": "mypassword" }
```

เหมือนกับ [01_POST_auth_register.md](./01_POST_auth_register.md) ทุกประการ

---

## Response `201`

```json
{
  "success": true,
  "data": {
    "id": "64b1f2c3d4e5f6a7b8c9d0e1",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "created_at": "2026-07-07T10:00:00Z"
  }
}
```

---

## Errors

| Code | Condition |
|------|-----------|
| `401` | Missing or invalid token |
| `409` | Email already exists |
| `422` | Validation failed |
| `500` | MongoDB write failed |
