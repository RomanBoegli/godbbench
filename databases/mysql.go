package databases

import (
	"database/sql"
	"fmt"
	"log"

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
		{Name: "inserts", Type: benchmark.TypeLoop, Stmt: "INSERT INTO GoBench.Generic (GenericId, Name, Balance, Description) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', {{call .RandInt63n 9999999999}}, '{{call .RandString 0 100 }}' );"},
		{Name: "selects", Type: benchmark.TypeLoop, Stmt: "SELECT * FROM GoBench.Generic WHERE GenericId = {{.Iter}};"},
		{Name: "updates", Type: benchmark.TypeLoop, Stmt: "UPDATE GoBench.Generic SET Name = '{{call .RandString 3 10 }}', Balance = {{call .RandInt63n 9999999999}} WHERE GenericId = {{.Iter}};"},
		{Name: "deletes", Type: benchmark.TypeLoop, Stmt: "DELETE FROM GoBench.Generic WHERE GenericId = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (m *Mysql) Setup() {
	if _, err := m.db.Exec("CREATE DATABASE IF NOT EXISTS GoBench;"); err != nil {
		log.Fatalf("failed to create database: %v\n", err)
	}
	if _, err := m.db.Exec("USE GoBench;"); err != nil {
		log.Fatalf("failed to USE GoBench: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS GoBench.Generic (GenericId INT PRIMARY KEY, Name VARCHAR(10), Balance DECIMAL, Description VARCHAR(100));"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := m.db.Exec("TRUNCATE GoBench.Generic;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (m *Mysql) Cleanup(closeConnection bool) {
	if _, err := m.db.Exec("DROP DATABASE IF EXISTS GoBench;"); err != nil {
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
	_, err := m.db.Exec(stmt)
	if err != nil {
		log.Printf("%v failed: %v", stmt, err)
	}
}
