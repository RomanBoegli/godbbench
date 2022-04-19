package databases

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/RomanBoegli/godbbench/benchmark"
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
		{Name: "inserts", Type: benchmark.TypeLoop, IterRatio: 1.0, Stmt: "INSERT INTO godbbench.Generic (GenericId, Name, Balance, Description) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', {{call .RandInt64}}, '{{call .RandString 0 100 }}' );"},
		{Name: "selects", Type: benchmark.TypeLoop, IterRatio: 1.0, Stmt: "SELECT * FROM godbbench.Generic WHERE GenericId = {{.Iter}};"},
		{Name: "updates", Type: benchmark.TypeLoop, IterRatio: 1.0, Stmt: "UPDATE godbbench.Generic SET Name = '{{call .RandString 3 10 }}', Balance = {{call .RandInt64}} WHERE GenericId = {{.Iter}};"},
		{Name: "deletes", Type: benchmark.TypeLoop, IterRatio: 1.0, Stmt: "DELETE FROM godbbench.Generic WHERE GenericId = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (p *Postgres) Setup() {
	if _, err := p.db.Exec("CREATE SCHEMA IF NOT EXISTS godbbench"); err != nil {
		log.Fatalf("failed to create schema: %v\n", err)
	}
	if _, err := p.db.Exec("CREATE TABLE IF NOT EXISTS godbbench.Generic (GenericId INT PRIMARY KEY, Name VARCHAR(10), Balance DECIMAL, Description VARCHAR(100));"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := p.db.Exec("TRUNCATE godbbench.Generic;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (p *Postgres) Cleanup(closeConnection bool) {
	if _, err := p.db.Exec("DROP TABLE IF EXISTS godbbench.Generic CASCADE;"); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}
	if _, err := p.db.Exec("DROP SCHEMA IF EXISTS godbbench CASCADE;"); err != nil {
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

	isInTransaciton := false
	singleStmts := strings.Split(stmt, ";")
	execTrans := []string{}
	for _, stmt := range singleStmts {

		stmt = strings.TrimSpace(stmt)

		if stmt == "BEGIN" {
			isInTransaciton = true
			continue
		}
		if stmt == "COMMIT" {
			isInTransaciton = false
			p.ExecTransaction(execTrans)
			execTrans = []string{}
			continue
		}

		if isInTransaciton {
			execTrans = append(execTrans, stmt)
		} else {
			p.ExecStatement(stmt)
		}
	}
}

// Exec executes the given statement on the database.
func (p *Postgres) ExecStatement(stmt string) {
	if stmt != "" {
		_, err := p.db.Exec(stmt)
		if err != nil {
			log.Printf("%v failed: %v", stmt, err)
		}
	}
}

// Exec executes the given statement on the database using transactions.
func (p *Postgres) ExecTransaction(singleStmts []string) {
	transaction, err := p.db.Begin()
	if err != nil {
		panic(err)
	}
	for _, stmt := range singleStmts {
		if stmt != "" {
			if _, err := transaction.Exec(stmt); err != nil {
				log.Fatalf("%v: failed(!): %v\n", stmt, err)
			}
		}
	}
	if err = transaction.Commit(); err != nil {
		log.Fatalf("%v: failed(!): %v\n", transaction, err)
	}
}
