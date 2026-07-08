# GET /api/users - List all users (paginated)

**← [_ROUTES.md](./_ROUTES.md)** | Spec: [FUNC §4](../FUNC_UserManagement.md#4-business-rules)

---

## Permission

Requires valid JWT.

---

## Query Params

| Param | Type | Required | Default | หมายเหตุ |
|-------|------|:--------:|---------|---------|
| `page` | int | - | `1` | ค่าที่ไม่ใช่จำนวนเต็มบวก fallback เป็น `1` |
| `limit` | int | - | `20` | สูงสุด `100` ต่อ request, ค่าที่ผิด fallback เป็น `20` |

spec ไม่ได้บังคับให้มี pagination แต่ใส่เพิ่มเพราะ list ทุก document โดยไม่จำกัดไม่เหมาะกับ production จริง

---

## Logic

1. Validate JWT -> `401` if invalid
2. Parse `page`/`limit` จาก query string, ค่าที่ผิด/ขาดหายใช้ default แทนการ error
3. `Find` documents ด้วย `Skip`/`Limit`, sort ตาม `_id` เพื่อให้ลำดับคงที่ข้าม page
4. `Count` documents ทั้งหมด (ไม่กรอง) สำหรับ `total`
5. Map เป็น UserResponse (exclude password)

---

## Response `200`

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "64b1f2c3d4e5f6a7b8c9d0e1",
        "name": "John Doe",
        "email": "john@example.com",
        "created_at": "2026-07-07T10:00:00Z"
      }
    ],
    "total": 42,
    "page": 1,
    "limit": 20
  }
}
```

Empty result: `items` เป็น `[]` ไม่ใช่ `null`, `total` เป็น `0`

---

## Errors

| Code | Condition |
|------|-----------|
| `401` | Missing or invalid token |
| `500` | MongoDB query failed |
