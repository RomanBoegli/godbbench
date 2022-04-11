package databases

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/RomanBoegli/gobench/benchmark"
)

// Mysql implements the bencher interface.
type Mysql struct {
	db *sql.DB
}

// NewMySQL returns a new mysql bencher.
func NewMySQL(host string, port int, user, password string, maxOpenConns int) *Mysql {
	if port == 0 {
		port = 3306
	}
	// username:password@protocol(address)/dbname?param=value
	dataSourceName := fmt.Sprintf("%v:%v@tcp(%v:%v)/", user, password, host, port)

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatalf("failed to open connection: %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	p := &Mysql{db: db}
	return p
}

// Benchmarks returns the individual benchmark functions for the mysql db.
func (m *Mysql) Benchmarks() []benchmark.Benchmark {
	return []benchmark.Benchmark{
		{Name: "inserts", Type: benchmark.TypeLoop, IterRatio: 1.0, Stmt: "INSERT INTO gobench.Generic (GenericId, Name, Balance, Description) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', {{call .RandInt63n 9999999999}}, '{{call .RandString 0 100 }}' );"},
		{Name: "selects", Type: benchmark.TypeLoop, IterRatio: 1.0, Stmt: "SELECT * FROM gobench.Generic WHERE GenericId = {{.Iter}};"},
		{Name: "updates", Type: benchmark.TypeLoop, IterRatio: 1.0, Stmt: "UPDATE gobench.Generic SET Name = '{{call .RandString 3 10 }}', Balance = {{call .RandInt63n 9999999999}} WHERE GenericId = {{.Iter}};"},
		{Name: "deletes", Type: benchmark.TypeLoop, IterRatio: 1.0, Stmt: "DELETE FROM gobench.Generic WHERE GenericId = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (m *Mysql) Setup() {
	if _, err := m.db.Exec("CREATE DATABASE IF NOT EXISTS gobench;"); err != nil {
		log.Fatalf("failed to create database: %v\n", err)
	}
	if _, err := m.db.Exec("USE gobench;"); err != nil {
		log.Fatalf("failed to USE gobench: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS gobench.Generic (GenericId INT PRIMARY KEY, Name VARCHAR(10), Balance DECIMAL, Description VARCHAR(100));"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := m.db.Exec("TRUNCATE gobench.Generic;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (m *Mysql) Cleanup(closeConnection bool) {
	if _, err := m.db.Exec("DROP DATABASE IF EXISTS gobench;"); err != nil {
		log.Printf("failed drop schema: %v\n", err)
	}
	if closeConnection {
		if err := m.db.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}
}

// Exec executes the given statement on the database.
func (m *Mysql) Exec(stmt string) {

	isInTransaciton := false
	singleStmts := strings.Split(stmt, ";")
	execTrans := []string{}
	for _, stmt := range singleStmts {

		stmt = strings.TrimSpace(stmt)

		if stmt == "START TRANSACTION" {
			isInTransaciton = true
			continue
		}
		if stmt == "COMMIT" {
			isInTransaciton = false
			m.ExecTransaction(execTrans)
			execTrans = []string{}
			continue
		}

		if isInTransaciton {
			execTrans = append(execTrans, stmt)
		} else {
			m.ExecStatement(stmt)
		}
	}
}

// Exec executes the given statement on the database.
func (m *Mysql) ExecStatement(stmt string) {

	if stmt != "" {
		_, err := m.db.Exec(stmt)
		if err != nil {
			log.Printf("%v failed: %v", stmt, err)
		}
	}

}

// Exec executes the given statement on the database using transactions.
func (m *Mysql) ExecTransaction(singleStmts []string) {
	transaction, err := m.db.Begin()
	if err != nil {
		panic(err)
	}
	for _, stmt := range singleStmts {
		if stmt != "" {
			if a, err := transaction.Exec(stmt); err != nil {
				log.Fatalf("%v: failed(!): %v\n%v\n", stmt, err, a)
			}
		}
	}
	if err = transaction.Commit(); err != nil {
		log.Fatalf("%v: failed(!): %v\n", transaction, err)
	}
}
