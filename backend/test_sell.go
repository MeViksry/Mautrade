package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://mautrade:mautrade@localhost:5432/mautrade?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Println("DB error:", err)
		return
	}
	defer pool.Close()

	var count int
	pool.QueryRow(context.Background(), "SELECT count(*) FROM layers").Scan(&count)
	fmt.Println("Total layers:", count)
}
