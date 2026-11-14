package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, os.Args[1])
	if err != nil {
		fmt.Println("FAIL:", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)
	var one int
	if err := conn.QueryRow(ctx, "select 1").Scan(&one); err != nil {
		fmt.Println("QUERY_FAIL:", err)
		os.Exit(2)
	}
	fmt.Println("OK")
}
