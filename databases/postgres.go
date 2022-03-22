package databases

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/RomanBoegli/gobench/benchmark"
)

// Postgres implements the bencher interface.
type Postgres struct {
	db *sql.DB
}

// NewPostgres returns a new postgres bencher.
func NewPostgres(host string, port int, user, password string, maxOpenConns int) *Postgres {
	if port == 0 {
		port = 5432
	}

	dataSourceName := fmt.Sprintf("host=%v port=%v user='%v' password='%v' sslmode=disable", host, port, user, password)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	db.SetMaxOpenConns(maxOpenConns)

	p := &Postgres{db: db}
	return p
}

// Benchmarks returns the individual benchmark statements for the postgres db.
func (p *Postgres) Benchmarks() []benchmark.Benchmark {
	return []benchmark.Benchmark{
		{Name: "inserts", Type: benchmark.TypeLoop, Stmt: "INSERT INTO GoBench.Generic (GenericId, Name, Balance, Description) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', {{call .RandInt63n 9999999999}}, '{{call .RandString 0 100 }}' );"},
		{Name: "selects", Type: benchmark.TypeLoop, Stmt: "SELECT * FROM GoBench.Generic WHERE GenericId = {{.Iter}};"},
		{Name: "updates", Type: benchmark.TypeLoop, Stmt: "UPDATE GoBench.Generic SET Name = '{{call .RandString 3 10 }}', Balance = {{call .RandInt63n 9999999999}} WHERE GenericId = {{.Iter}};"},
		{Name: "deletes", Type: benchmark.TypeLoop, Stmt: "DELETE FROM GoBench.Generic WHERE GenericId = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (p *Postgres) Setup() {
	if _, err := p.db.Exec("CREATE SCHEMA IF NOT EXISTS GoBench"); err != nil {
		log.Fatalf("failed to create schema: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS GoBench.Generic (GenericId INT PRIMARY KEY, Name VARCHAR(10), Balance DECIMAL, Description VARCHAR(100));"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := p.db.Exec("TRUNCATE GoBench.Generic;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Postgres) Cleanup(closeConnection bool) {
	if _, err := p.db.Exec("DROP TABLE IF EXISTS GoBench.Generic CASCADE;"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP SCHEMA IF EXISTS GoBench CASCADE;"); err != nil {
		log.Printf("failed drop schema: %v\n", err)
	}
	if closeConnection {
		if err := p.db.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}
}

// Exec executes the given statement on the database.
func (p *Postgres) Exec(stmt string) {
	_, err := p.db.Exec(stmt)
	if err != nil {
		log.Printf("%v failed: %v", stmt, err)
	}
}
