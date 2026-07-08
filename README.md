# 7solutions Backend Challenge

This repo has two parts:

- **[`user-management-api/`](./user-management-api)** — Section 1: Go + Gin + MongoDB + JWT User Management API (code + tests)
- **[`lottery-search-design/`](./lottery-search-design)** — Section 2: Lottery Search System design proposal. Read it at [`LOTTERY_DESIGN.md`](./lottery-search-design/LOTTERY_DESIGN.md), or for an easier read open [`LOTTERY_DESIGN_VISUAL.html`](./lottery-search-design/LOTTERY_DESIGN_VISUAL.html) in a browser.

---

# User Management API

REST API สำหรับ register, login, และ CRUD user เขียนด้วย Go (Gin), MongoDB, และ JWT (HS256) ออกแบบตาม Hexagonal Architecture (ports & adapters)

> All commands below are run from inside `user-management-api/` (`cd user-management-api` first).

## Architecture

```
Handler (HTTP adapter) -> Service (business logic) -> UserRepository interface (port) -> MongoUserRepository (adapter)
```

Application layer (`internal/application`) พึ่งพาแค่ interface `ports.UserRepository` เท่านั้น ไม่รู้จัก MongoDB driver โดยตรง
ผลคือ business logic สามารถ unit test ด้วย in-memory fake repository ได้เลย - ไม่ต้องมี database จริง - และเปลี่ยน storage technology ได้โดยไม่ต้องแตะ service method แม้แต่ตัวเดียว

## Project Structure

```
user-management-api/
├── cmd/server/main.go          entry point, wiring, graceful shutdown
├── config/config.go            env var loading
├── internal/
│   ├── domain/                 User struct, Go ล้วน ไม่มี framework dependency
│   ├── ports/                  UserRepository interface
│   ├── application/            AuthService, UserService (business logic)
│   └── adapters/
│       ├── mongo/              MongoUserRepository (implement port)
│       └── http/                handlers, middleware, router
└── pkg/
    ├── jwt/                    sign + validate HS256 token
    ├── password/                bcrypt hash + compare
    └── response/                 JSON envelope มาตรฐาน
```

## Setup

### รันแบบ Local (ไม่ใช้ Docker)

ต้องมี Go 1.26+ และ MongoDB ที่รันอยู่

```bash
cp .env.example .env
# แก้ .env ถ้า MongoDB ไม่ได้อยู่ที่ localhost:27017

go mod download
go run ./cmd/server
```

Server จะรันที่ `http://localhost:8080` (หรือตาม `PORT` ที่กำหนด)

### Docker Compose

```bash
cp .env.example .env
docker-compose up --build
```

คำสั่งนี้จะเปิดทั้ง API container และ MongoDB container พร้อมต่อกันในเครือข่าย Compose เดียวกัน

### ทดสอบด้วย Postman

มี Postman collection ตัวอย่างให้พร้อมใช้ที่ `postman/7solutions-backend.postman_collection.json`
เปิด Postman -> **File -> Import** -> เลือกไฟล์นี้ได้เลย

ลำดับที่แนะนำ: รัน `docker-compose up --build` ให้ API ขึ้นก่อน แล้ว import ไฟล์นี้ -> กด **Register** ->
กด **Login** (จะเก็บ token ให้อัตโนมัติ) -> ที่เหลือทุก request ใช้ token/userId ที่เก็บไว้ให้เองโดยไม่ต้อง copy-paste เอง

## Running Tests

```bash
go test ./...
```

Unit test ของ `AuthService` และ `UserService` ใช้ in-memory fake ที่ implement `ports.UserRepository`
ไม่ต้องต่อ MongoDB จริงตอนรัน test

## JWT Guide

1. Register user:

   ```bash
   curl -X POST http://localhost:8080/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{"name":"John Doe","email":"john@example.com","password":"secret123"}'
   ```

2. Login เพื่อรับ token:

   ```bash
   curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"john@example.com","password":"secret123"}'
   ```

   Response:

   ```json
   { "success": true, "data": { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." } }
   ```

3. คัดลอกค่า `token` แล้วส่งเป็น Bearer token ทุกครั้งที่เรียก protected endpoint:

   ```bash
   curl http://localhost:8080/api/users \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
   ```

Token หมดอายุ 24 ชั่วโมงหลังออก ไม่มี refresh-token endpoint - ต้อง login ใหม่เพื่อรับ token ใหม่

## Endpoints

ทุก response ใช้ envelope เดียวกัน:

```json
{ "success": true, "data": { ... } }
{ "success": false, "error": { "code": "NOT_FOUND", "message": "user not found" } }
```

### POST /api/auth/register (public)

Request:

```json
{ "name": "John Doe", "email": "john@example.com", "password": "secret123" }
```

Response `201`:

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

Errors: `400` ขาด field, `409` email ซ้ำในระบบ, `422` email ผิด format หรือ password สั้นกว่า 6 ตัว

### POST /api/auth/login (public)

Request:

```json
{ "email": "john@example.com", "password": "secret123" }
```

Response `200`:

```json
{ "success": true, "data": { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." } }
```

Errors: `400` ขาด field, `401` invalid credentials (ข้อความเดียวกันไม่ว่า email จะไม่มีในระบบ หรือ password ผิด)

### POST /api/users (protected)

Logic เดียวกับ `POST /api/auth/register` ทุกอย่าง (validate, hash password, check email ซ้ำ) ต่างกันแค่ route นี้ต้องมี JWT
เปิดไว้ให้ "Create a new user" ที่อยู่ในกลุ่ม User Operations (CRUD ที่ protect ด้วย token ทั้งหมด) มี route ของตัวเอง แยกจาก self-service register ที่เป็น public

Request:

```json
{ "name": "Jane Doe", "email": "jane@example.com", "password": "mypassword" }
```

Response `201`:

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

Errors: `401` ไม่มี token, `409` email ซ้ำในระบบ, `422` validation ไม่ผ่าน

### GET /api/users (protected, paginated)

Query params: `page` (default `1`), `limit` (default `20`, สูงสุด `100`) - ค่าที่ผิดหรือขาดหายใช้ default แทนการ error

Response `200`:

```json
{
  "success": true,
  "data": {
    "items": [
      { "id": "64b1f2c3d4e5f6a7b8c9d0e1", "name": "John Doe", "email": "john@example.com", "created_at": "2026-07-07T10:00:00Z" }
    ],
    "total": 42,
    "page": 1,
    "limit": 20
  }
}
```

`items` เป็น `[]` (ไม่ใช่ `null`) เมื่อไม่มี user เลย

### GET /api/users/:id (protected)

Response `200`:

```json
{
  "success": true,
  "data": { "id": "64b1f2c3d4e5f6a7b8c9d0e1", "name": "John Doe", "email": "john@example.com", "created_at": "2026-07-07T10:00:00Z" }
}
```

Errors: `400` ObjectID ไม่ถูกต้อง, `404` ไม่พบ user

### PUT /api/users/:id (protected)

Request (ต้องมีอย่างน้อย 1 field):

```json
{ "name": "John Updated", "email": "john.updated@example.com" }
```

Response `200`:

```json
{
  "success": true,
  "data": { "id": "64b1f2c3d4e5f6a7b8c9d0e1", "name": "John Updated", "email": "john.updated@example.com", "created_at": "2026-07-07T10:00:00Z" }
}
```

Errors: `400` ObjectID ไม่ถูกต้อง หรือไม่ส่ง field มาเลย, `404` ไม่พบ user, `409` email ใหม่ถูกใช้โดย user อื่นแล้ว

### DELETE /api/users/:id (protected)

Response `200`:

```json
{ "success": true, "data": { "message": "user deleted successfully" } }
```

Errors: `400` ObjectID ไม่ถูกต้อง, `404` ไม่พบ user

## Assumptions & Design Decisions

- **Register กับ Create User ใช้ logic เดียวกัน** spec แบ่ง "User registration" (หมวด Authentication) กับ "Create a new user" (หมวด User Operations ที่ protect ด้วย JWT ทั้งหมด) เป็นคนละ requirement ผมจึงเปิด 2 route: `POST /api/auth/register` (public, self-service) และ `POST /api/users` (ต้องมี JWT) ทั้งคู่เรียก handler เดียวกัน ไม่มี business logic ซ้ำ
- **Hard delete** `DELETE /api/users/:id` ลบ User ทิ้งถาวร ไม่ได้เป็น soft-delete
- **เพิ่ม pagination บน list** `GET /api/users` เดิม list ทุก user โดยไม่จำกัด ซึ่งถ้า user ในระบบมีจำนวนมากจะเป็นปัญหา ผมเลยเพิ่ม `page`/`limit` (default 1/20, สูงสุด 100) เข้าไปเพราะคิดว่าเป็นฟีเจอร์ที่สมควรต้องมีสำหรับ production จริง ค่าที่ผิดจะ fallback เป็น default แทนการคืน `400` เพื่อไม่ให้ validation error เข้มงวดเกินไปสำหรับฟีเจอร์เสริมนี้
- **401 message เดียวกันทั้งสองกรณี login ผิด** ไม่ว่า email จะไม่มีในระบบ หรือ password ผิด API จะคืนข้อความ `invalid credentials` เหมือนกัน ถ้าแยกข้อความจะเปิดช่องให้ผู้โจมตีเดาได้ว่า email ไหนมีอยู่ในระบบจริง
- **400 vs 422 ตอน register/update** field ที่ required แต่ขาดหาย = `400` ส่วน field ที่ส่งมาแล้วแต่ผิด format (เช่น email ผิด, password สั้นกว่า 6 ตัว) = `422` ทั้งสองเคสมาจาก validator library ตัวเดียวกัน handler จะตรวจว่า validation tag ไหนที่ fail (`required` หรือ tag อื่น) เพื่อเลือก status code ให้ถูก
- **Hexagonal architecture** Application layer พึ่งพาแค่ interface `ports.UserRepository` ไม่รู้จัก MongoDB driver โดยตรง ทำให้ business logic ทดสอบได้โดยไม่ต้องมี database จริง และเปลี่ยน storage technology ได้

---

# Lottery Search System (Design Only)

รายละเอียดเต็มอยู่ที่ [`lottery-search-design/LOTTERY_DESIGN.md`](./lottery-search-design/LOTTERY_DESIGN.md) (ไม่มี coding ตาม requirement) หรือถ้าอยากอ่านง่ายๆ แบบมีตัวอย่างโต้ตอบได้ เปิดไฟล์ [`lottery-search-design/LOTTERY_DESIGN_VISUAL.html`](./lottery-search-design/LOTTERY_DESIGN_VISUAL.html) ในเบราว์เซอร์ได้เลย สรุปสั้นๆ ตามคำถามที่โจทย์ถาม:

**Database?** เลือก **PostgreSQL** เพราะมีคำสั่ง `SELECT FOR UPDATE SKIP LOCKED` ที่แก้ปัญหา "สองคนแย่ง ticket ใบเดียวกัน" ได้ตรงจุด โดยไม่ต้องเขียนโค้ดกันชนเอง

**Search Algorithm?** แยกเก็บตัวเลขแต่ละหลักของ ticket ไว้คนละคอลัมน์ (`d0`-`d5`) แล้วสร้าง index ให้ทุกคอลัมน์ ทำให้ pattern ที่รู้บางหลักค้นเจอเร็วโดยไม่ต้องสร้างโครงสร้างข้อมูลพิเศษอย่าง Trie หรือ Bitmap เพิ่มเอง

**Concurrency (กัน race condition)?** ใช้ `SELECT ... FOR UPDATE SKIP LOCKED` ล็อก ticket ทันทีตอนค้นเจอ คนที่มาทีหลังจะถูกข้ามไปหาใบถัดไปให้อัตโนมัติ พร้อมระบบจองชั่วคราวที่หมดอายุใน 5 นาที กัน ticket หายจาก pool ถ้าค้นแล้วไม่ซื้อ

**Performance?** ยิ่งรู้ตัวเลขในหลักไหนแน่นอนมากเท่าไหร่ ก็ยิ่งค้นเจอเร็วขึ้น เพราะ Postgres ใช้ index ของหลักนั้นตัดตัวเลือกทิ้งได้ทันที เลือกใช้ B-tree index เพราะเร็วพอสำหรับ use case นี้และกินหน่วยความจำน้อยกว่าทางเลือกอย่าง Trie มาก

ดูฉบับเต็ม + ทางเลือกอื่นที่พิจารณาแล้ว (Redis, MongoDB, Elasticsearch, ClickHouse) และ trade-off ทั้งหมดได้ที่ [`lottery-search-design/LOTTERY_DESIGN.md`](./lottery-search-design/LOTTERY_DESIGN.md) หรือดูฉบับ interactive ที่ [`lottery-search-design/LOTTERY_DESIGN_VISUAL.html`](./lottery-search-design/LOTTERY_DESIGN_VISUAL.html)
