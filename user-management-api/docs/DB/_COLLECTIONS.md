# MongoDB Collections - User Management

**Spec:** [FUNC_UserManagement.md](../FUNC_UserManagement.md)  
**Database:** `7solutions`

---

## Collection Index

| # | Collection | Description |
|---|------------|-------------|
| 01 | `users` | User accounts |

---

## Collection: `users`

### Document Shape

```json
{
  "_id":        ObjectId("64b1f2c3d4e5f6a7b8c9d0e1"),
  "name":       "John Doe",
  "email":      "john@example.com",
  "password":   "$2a$10$...",
  "created_at": ISODate("2026-07-07T10:00:00Z")
}
```

### Go Struct

```go
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `bson:"name"`
    Email     string             `bson:"email"`
    Password  string             `bson:"password"`
    CreatedAt time.Time          `bson:"created_at"`
}
```

### Indexes

| Index | Fields | Options | Purpose |
|-------|--------|---------|---------|
| `_id` (default) | `_id` | unique | primary key |
| `idx_users_email` | `email` | unique | enforce uniqueness + login lookup |

```go
// Create indexes on startup
indexModels := []mongo.IndexModel{
    {
        Keys:    bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    },
}
collection.Indexes().CreateMany(ctx, indexModels)
```

### Constraints

- `email` unique enforced at both application layer AND database index
- `password` is always bcrypt hash - never plaintext
- No soft delete - `DeleteOne` removes the document permanently
- No `updated_at` field - spec only mentions `CreatedAt`

---

## Shared Context

### Connection Setup

```go
client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
db := client.Database(cfg.MongoDB)
usersCollection := db.Collection("users")
```

### Context Propagation

All MongoDB operations use `context.Context` passed from the handler:
```go
func (r *MongoUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
```

This enables request-scoped timeouts and cancellation propagation.
