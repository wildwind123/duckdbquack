## DuckDB via Quack: always use quack_query

The database is a remote DuckDB 1.5.5 server reached over Quack. Use `@duckdb/node-api@1.5.5-r.x` (match the server's 1.5.5).

Setup, once per connection (no ATTACH needed):

```js
await con.run("LOAD quack");   // INSTALL quack once per machine
await con.run("CREATE SECRET (TYPE quack, TOKEN $1)", [process.env.QUACK_TOKEN]);
```

Run EVERY statement through `quack_query`, via this one helper:

```js
const QUACK_ADDR = process.env.QUACK_ADDR;   // 'quack:localhost:9494'
const lit = (v) => v === null || v === undefined ? 'NULL'
  : typeof v === 'number' || typeof v === 'bigint' ? String(v)
  : typeof v === 'boolean' ? (v ? 'TRUE' : 'FALSE')
  : `'${String(v).replaceAll("'", "''")}'`;
const query = async (sql) =>
  (await con.runAndReadAll("FROM quack_query($1, $2)", [QUACK_ADDR, sql])).getRowObjectsJson();

await query(`SELECT * FROM users WHERE id = ${lit(id)}`);
await query(`INSERT INTO users VALUES (${lit(id)}, ${lit(name)})`);
await query(`UPDATE users SET name = ${lit(name)} WHERE id = ${lit(id)}`);
```

Rules:
- Never use `ATTACH` / `remote.<table>`: UPDATE/DELETE fail on it, and WHERE/aggregates/joins run client-side, downloading the whole table (1M rows: 61.7 ms vs 2.6 ms).
- Table names have no `remote.` prefix; the SQL runs on the server.
- Every value in the SQL goes through `lit()`. Never interpolate raw input: bound params don't reach the inner SQL, and the server accepts multiple statements per call, so unescaped input can inject extra statements.
- One statement per call. Batch lookups (`WHERE id IN (...)`) instead of looping queries.
- Retry writes on a transaction conflict error.