# duckdbquack

Go server exposing a DuckDB database over the Quack protocol (`go/main.go`), plus a Go client (`go/client/main.go`). DuckDB version: 1.5.5 (duckdb-go `v2.10505.0`); clients should use a matching 1.5.5 version.

## Querying through Quack: always use `quack_query`

Clients connect with (no `ATTACH` needed):

```sql
INSTALL quack; LOAD quack;
CREATE SECRET (TYPE quack, TOKEN '<token>');
```

Run EVERY statement (`SELECT`, `INSERT`, `UPDATE`, `DELETE`, `CREATE TABLE`, ...) through `quack_query`, so it executes entirely on the server:

```sql
FROM quack_query('quack:localhost:9494', 'SELECT * FROM users WHERE id = 2');
FROM quack_query('quack:localhost:9494', 'UPDATE users SET name = ''bob'' WHERE id = 2');
```

Do not use `ATTACH ... AS remote` / `remote.<table>`:
- `UPDATE` and `DELETE` fail on it (`Binder Error: Can only update base table`).
- `WHERE`, aggregates and joins run on the client, so the whole table is downloaded. Measured on 1M rows: point lookup 61.7 ms via `remote.` vs 2.6 ms via `quack_query`.

Rules for `quack_query`:
- The first argument is the `quack:` address, not an alias.
- Table names have no `remote.` prefix; the SQL runs on the server.
- The inner SQL does not support bound parameters. Every value put into it must be escaped as a SQL literal (double single quotes: `''bob''`; in Go use `quote()`). The server accepts multiple statements per call, so unescaped input can inject extra statements.
- One statement per call.
- Retry writes on a transaction conflict error.

## Concurrency

Only the server process opens `data.duckdb`; other processes can't open the file while it runs. Many clients can read and write through Quack at once.

## Testing

Test against a throwaway server, never the real `data.duckdb`:

```sh
QUACK_TOKEN=super_secret ./duckdbquack -db :memory: -listen quack:localhost:9495
./quack-client -addr quack:localhost:9495 -q "..."
```
