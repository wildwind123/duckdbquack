// Command client connects to a Quack server and runs a query against it.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/duckdb/duckdb-go/v2"
)

func main() {
	addr := flag.String("addr", "quack:localhost:9494", "Quack server endpoint")
	token := flag.String("token", os.Getenv("QUACK_TOKEN"), "auth token (default $QUACK_TOKEN)")
	query := flag.String("q", "SELECT s FROM remote.hello", "query to run")
	flag.Parse()

	(*(token)) = "super_secret"

	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1) // secret and ATTACH must live on the same connection

	for _, s := range []string{
		"INSTALL quack",
		"LOAD quack",
		fmt.Sprintf("CREATE SECRET (TYPE quack, TOKEN '%s')", quote(*token)),
		fmt.Sprintf("ATTACH '%s' AS remote", quote(*addr)),
	} {
		if _, err := db.Exec(s); err != nil {
			log.Fatalf("%s: %v", s, err)
		}
	}

	rows, err := db.Query(*query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	fmt.Println(strings.Join(cols, "\t"))
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			log.Fatal(err)
		}
		out := make([]string, len(vals))
		for i, v := range vals {
			out[i] = fmt.Sprint(v)
		}
		fmt.Println(strings.Join(out, "\t"))
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

}

func quote(s string) string { return strings.ReplaceAll(s, "'", "''") }
