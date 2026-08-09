# Specific instructions — Database and persistence

Complements `AGENTS.md`.

## ORM and driver

- GORM over Postgres is the standard for a general relational domain. Use `github.com/raykavin/gobox/database/gorm` as the connection factory (built-in connection retry, configurable pool via `GormConfig`) — don't open the connection manually with `gorm.Open` directly.
- SQLite is acceptable for local development or low-scale services without shared database infrastructure.
- Plain pgx (`jackc/pgx/v5`) without an ORM is acceptable **only** for simple entities (few columns, no deep relations) with high performance/explicit-SQL-control pressure. For entities with many nested JSONB fields, this approach becomes noticeably more verbose than GORM — don't choose plain pgx purely for stylistic preference when the entity has that shape; weigh the real trade-off first.

## Base model and entity ID

Recommended default: `uint` autoincrement, embedded via `BaseModel`. Use `uuid.UUID` (`gorm:"default:gen_random_uuid()"`) when the ID needs to be exposed or generated externally without coordination (e.g., multi-service sync, client-side generation).

Recommended structure (regardless of the chosen ID type):
```go
type BaseModel struct {
    ID        <uint or uuid.UUID>
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```
Embed it in every persisted entity. For join/append-only history entities that don't need update/soft-delete, it's acceptable to omit `DeletedAt` — but document why in the code itself.

## Soft delete

- `gorm.DeletedAt` is the standard mechanism, embedded via `BaseModel`.
- If a column needs uniqueness that ignores soft-deleted rows, use a partial unique index: `gorm:"uniqueIndex:uq_x,where:deleted_at IS NULL"` — without this protection it's not possible to recreate a record with the same unique value after a soft delete.

## Migrations vs. AutoMigrate

- `db.AutoMigrate(...)` called once at process boot is convenient for an early stage/prototype/low-risk domain, but has real risk: no version history, no `Down`, no human review before altering production, and index creation doesn't run `CONCURRENTLY` (can lock a large table).
- For critical/financial data: prefer versioned SQL migrations (`golang-migrate`, `NNNNNN_description.up.sql`/`.down.sql` format) applied by a dedicated `cmd/migrator` binary, **not** run automatically at API/worker boot.
- **Non-negotiable rule regardless of the above decision**: don't modify a migration already applied in production — create a new one. Any `DROP COLUMN`/`DROP TABLE`/`ALTER TYPE`/`RENAME` on a table with real data requires explicit human approval (see `AGENTS.md`).

## Transactions and Unit of Work

- Implement a `port.TransactionRunner` interface with `RunInTransaction(ctx, fn func(ctx) error)`, over `gorm.DB.Transaction`, propagated via `context.WithValue` — each repository uses a `dbFor(ctx, r.db)` helper that uses the ambient transaction if present, otherwise the base connection.
- Multi-step usecases (e.g., two sequential `Upsert`s across different tables) need to be atomic — don't write a multi-step usecase without a transaction.
- For repository tests involving a real transaction, use SQLite in-memory (`sqlite.Open(":memory:")` or `file:name?mode=memory&cache=shared`) instead of mocking every collaborator.

## N+1 and queries

- Use GORM's `Preload("Relation")` to hydrate associations in read methods. Chain (`Preload("A.B")`) for second-level relations.
- Reserve manual `.Joins(...)` for search filters (e.g., `Joins("LEFT JOIN people ON ...")` inside a `Filter`/`Search` method), never to load associations that should use `Preload`.

## Pagination at the repository layer

Via `github.com/raykavin/gobox/pagination` (`Scope(query, &total)` as a `gorm.Scopes`). See `instructions/api.md` for the full end-to-end usage pattern.

## Write idempotency

- Preferred pattern for "find or create" (when the caller needs the record back, existing or new): content hash + composite unique index + find-or-create in the usecase.
- Preferred pattern for "write and ignore duplicate" (when the caller doesn't need the record back): GORM's `clause.OnConflict{DoNothing: true}` or `{Columns: [...], DoNothing: true}`.
- Preferred pattern for a full upsert (create or overwrite): `clause.OnConflict{Columns: [...], UpdateAll: true}`.
- Don't use GORM's `FirstOrCreate` without evaluating the race condition — prefer the explicit find-then-create-or-conflict pattern above.

## Indexes, FKs, and naming

- Index naming convention: `idx_<table>_<column(s)>`. Composite uniques: prefer `uq_<table>_<columns>` — avoid the `idx_` prefix for unique indexes.
- FKs via GORM: `gorm:"foreignKey:X;references:ID"`. Explicitly document when a reference is intentionally without an FK ("weak reference" documented in the migration comment and the Go struct) — don't add a "corrective" FK without understanding why it was omitted.
- When writing SQL migrations (when the project uses that mechanism): group `CREATE INDEX`/`ALTER TABLE ... ADD CONSTRAINT` in idempotent blocks (`DO $$ IF NOT EXISTS ... END$$;`). Use `CREATE INDEX CONCURRENTLY` when adding an index to a table that already has data in production.

## Risks that require human approval

`DROP TABLE`, `DROP COLUMN`, `ALTER TYPE`, `RENAME TABLE`/`RENAME COLUMN` in any migration that could run against real data.
