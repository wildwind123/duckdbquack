// Command duckdbquack serves a DuckDB database over the Quack client/server protocol.
//
// Clients connect from any DuckDB (CLI, Python, ...) with:
//
//	CREATE SECRET (TYPE quack, TOKEN '<token>');
//	ATTACH 'quack:localhost' AS remote;
//	FROM remote.hello;
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/duckdb/duckdb-go/v2"
)

func main() {
	dbPath := flag.String("db", "data.duckdb", "DuckDB database file (use ':memory:' for in-memory)")
	listen := flag.String("listen", "quack:localhost:9494", "Quack endpoint to listen on")
	token := flag.String("token", os.Getenv("QUACK_TOKEN"), "auth token clients must present (default $QUACK_TOKEN)")
	flag.Parse()

	if *token == "" {
		log.Fatal("a token is required: pass -token or set QUACK_TOKEN")
	}

	dsn := *dbPath
	if dsn == ":memory:" {
		dsn = ""
	}
	db, err := sql.Open("duckdb", dsn)
	if err != nil {
		log.Fatalf("open duckdb: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	// Keep one connection open for the lifetime of the process so the
	// database instance (and the server running inside it) stays alive.
	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	stmts := []string{
		"INSTALL quack",
		"LOAD quack",
		"CREATE TABLE IF NOT EXISTS hello AS FROM VALUES ('world') v(s)",
		fmt.Sprintf("CALL quack_serve('%s', token = '%s')", quote(*listen), quote(*token)),
	}
	for _, s := range stmts {
		if _, err := conn.ExecContext(ctx, s); err != nil {
			log.Fatalf("%s: %v", strings.SplitN(s, "(", 2)[0], err)
		}
	}

	log.Printf("serving %s on %s (Ctrl+C to stop)", *dbPath, *listen)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Print("shutting down")
}

// quote escapes a value for use inside a single-quoted SQL string literal.
func quote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
