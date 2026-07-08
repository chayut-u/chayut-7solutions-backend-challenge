# Lottery Search System - Design Proposal

**Type:** Design document only (No code implementation)

---

## 1. Overview

This system manages a pool of 1 million pre-generated lottery tickets. Users can search tickets using a 6-character pattern where `*` represents any digit.

Examples:

| Pattern | Description |
|---------|--------------|
| `123***` | Starts with 123 |
| `****23` | Ends with 23 |
| `1****5` | Starts with 1 and ends with 5 |

The main challenge is not searching the tickets, but making sure the same ticket is never returned to two users at the same time.

To solve this, I would use:

- PostgreSQL
- Per-digit indexes for searching
- `SELECT ... FOR UPDATE SKIP LOCKED`
- Reservation with expiration

This keeps the solution simple, reliable, and suitable for production.

---

## 2. System Architecture

```
                User
                  │
                  ▼
            Search API
                  │
                  ▼
          Pattern Parser
                  │
                  ▼
            PostgreSQL
      - Search tickets
      - Reserve ticket
      - Row locking
                  │
                  ▼
            Return Result
```

The API is stateless, so multiple API servers can be added easily if traffic increases.

All concurrency control is handled by PostgreSQL, so there is no need for Redis locks or other distributed locking systems.

---

## 3. Database Choice

I would choose PostgreSQL.

Reasons:

- Supports ACID transactions
- Excellent indexing performance
- Built-in row locking
- Easy to maintain
- Mature and widely used

The biggest reason is support for

```sql
SELECT ... FOR UPDATE SKIP LOCKED
```

This allows multiple users to search at the same time without returning the same ticket.

Other databases were considered:

| Database | Reason not chosen |
|----------|---------------------|
| MongoDB | Flexible, but wildcard searching is less efficient. |
| Redis | Extremely fast, but requires additional logic for searching and reservation. |
| Elasticsearch | Excellent search engine, but not designed for transactional ticket allocation. |

---

## 4. Data Model

Each ticket contains:

| Field | Description |
|-------|--------------|
| `id` | Primary key |
| `number` | Original 6-digit ticket |
| `d0`-`d5` | Individual digits used for searching |
| `status` | available / reserved / sold |
| `reserved_until` | Reservation expiration time |

Example

| number | d0 | d1 | d2 | d3 | d4 | d5 |
|--------|----|----|----|----|----|----|
| 123456 | 1 | 2 | 3 | 4 | 5 | 6 |

Each digit column has its own index.

This allows the database to filter only the required digits instead of scanning all tickets.

---

## 5. Search Algorithm

When a search request arrives, the pattern is converted into SQL conditions.

Example

Pattern

```
1****5
```

becomes

```sql
WHERE d0='1'
AND d5='5'
AND status='available'
```

Another example

Pattern

```
****23
```

becomes

```sql
WHERE d4='2'
AND d5='3'
AND status='available'
```

Only the specified digits are included in the query.

For example, if the pattern is `1****5`, only the first and last digits are checked. Wildcards are ignored.

This approach supports wildcards in any position without requiring complex data structures.

---

## 6. Ticket Allocation

Searching alone is not enough.

The system must also make sure the same ticket is not given to two users.

I would perform searching and reservation inside a single transaction.

```sql
SELECT ...
FOR UPDATE SKIP LOCKED
```

The transaction begins, runs this search-and-lock query, updates the ticket's status to reserved, and then commits.

If another user searches at the same time, PostgreSQL skips the locked row and returns the next available ticket.

This guarantees that duplicate allocation cannot happen.

---

## 7. Reservation Flow

Instead of marking a ticket as sold immediately, I would use three states.

```
Available
     │
     ▼
Reserved
     │
     ├── Purchase completed
     ▼
Sold

Reserved
     │
     └── Expired
          ▼
Available
```

A reservation lasts for 5 minutes.

If the user leaves or never completes the purchase, the ticket automatically becomes available again.

This prevents tickets from being permanently locked.

---

## 8. Performance

Searching 1 million tickets is not difficult if indexes are used correctly.

Most searches only filter a few indexed columns, so PostgreSQL can find a matching ticket quickly without scanning the entire table.

Since the query always returns only one ticket (`LIMIT 1`), the database stops searching as soon as a matching ticket is found.

This keeps response time low even with a large dataset.

---

## 9. Trade-offs

**Advantages**

- Simple architecture
- Easy to maintain
- Strong transaction support
- No duplicate ticket allocation
- Easy to scale API servers horizontally

**Limitations**

- More indexes require additional storage.
- Searching with `******` is slower because there are no filtering conditions.
- PostgreSQL may experience lock contention under extremely high traffic, although this is acceptable for a dataset of 1 million tickets.

---

## 10. Future Improvements

If the system grows much larger, possible improvements include:

- Cache frequently searched patterns to reduce database load.
- Partition the ticket table when data grows beyond tens of millions of rows.
- Add read replicas for reporting and analytics.
- Monitor search patterns to optimize indexes based on real usage.

---

## Conclusion

This design focuses on three main goals:

- Fast wildcard searching using indexed digit columns.
- Safe ticket allocation using PostgreSQL row locking.
- A simple architecture that is easy to operate and scale.

For a dataset of 1 million lottery tickets, PostgreSQL with indexed digit columns and `SELECT ... FOR UPDATE SKIP LOCKED` provides a practical, reliable, and production-ready solution while keeping the implementation straightforward.

The solution is simple to implement, supports concurrent users safely, and can be scaled as the system grows.
