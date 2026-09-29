# duckdbquack

A small Go server that exposes a DuckDB database over the
[Quack](https://github.com/duckdb/duckdb-quack) client/server protocol (the protocol runs over HTTP, default port 9494).

## Run the server

```sh
go build -o duckdbquack .
QUACK_TOKEN=super_secret ./duckdbquack -db data.duckdb -listen quack:localhost:9494
```

Flags:
- `-db`: the database file. Use `:memory:` for an in-memory database.
- `-listen`: the Quack endpoint. Use `quack:0.0.0.0:9494` with `-allow-other-hostname` to accept remote clients.
- `-token`: the auth token. Defaults to `$QUACK_TOKEN`.
- `-allow-other-hostname`: allow `-listen` on a hostname other than `localhost`. Off by default.

On first start the server creates a demo table `hello`.

## Connect

From the DuckDB CLI (or any DuckDB client):

```sql
CREATE SECRET (TYPE quack, TOKEN 'super_secret');
ATTACH 'quack:localhost' AS remote;
FROM remote.hello;
CREATE TABLE remote.t2 AS SELECT 42 AS answer;
```

Or with the included Go client:

```sh
go build -o quack-client ./client
QUACK_TOKEN=super_secret ./quack-client -q "FROM remote.hello"
```

## Notes

- Quack is an experimental DuckDB extension. It is installed automatically on first run, so the server needs internet access the first time.
- If `HTTP_PROXY` is set, add localhost to `NO_PROXY`. Otherwise clients fail with `Bad Gateway`.
