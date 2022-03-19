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
		{Name: "inserts", Type: benchmark.TypeLoop, Stmt: "INSERT INTO dbbench.simple (id, name, balance, description) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', {{call .RandInt63n 9999999999}}, '{{call .RandString 0 100 }}' );"},
		{Name: "selects_Id", Type: benchmark.TypeLoop, Stmt: "SELECT * FROM dbbench.simple WHERE id = {{.Iter}};"},
		{Name: "selects_Text", Type: benchmark.TypeLoop, Stmt: "SELECT * FROM dbbench.simple WHERE description LIKE '%{{call .RandString 0 5}}%';"},
		{Name: "updates", Type: benchmark.TypeLoop, Stmt: "UPDATE dbbench.simple SET name = '{{call .RandString 3 10 }}', balance = {{call .RandInt63n 9999999999}} WHERE id = {{.Iter}};"},
		//{Name: "deletes", Type: benchmark.TypeLoop, Stmt: "DELETE FROM dbbench.simple WHERE id = {{.Iter}};"},
		//{Name: "relation_insert0", Type: benchmark.TypeLoop, Stmt: "INSERT INTO dbbench.relational_one (oid, balance_one) VALUES( {{.Iter}}, {{call .RandInt63n 9999999999}});"},
		//{Name: "relation_insert1", Type: benchmark.TypeLoop, Stmt: "INSERT INTO dbbench.relational_two (relation, balance_two) VALUES( {{.Iter}}, {{call .RandInt63n 9999999999}});"},
		//{Name: "relation_select", Type: benchmark.TypeLoop, Stmt: "SELECT * FROM dbbench.relational_two INNER JOIN dbbench.relational_one ON relational_one.oid = relational_two.relation WHERE relation = {{.Iter}};"},
		//{Name: "relation_delete1", Type: benchmark.TypeLoop, Stmt: "DELETE FROM dbbench.relational_two WHERE relation = {{.Iter}};"},
		//{Name: "relation_delete0", Type: benchmark.TypeLoop, Stmt: "DELETE FROM dbbench.relational_one WHERE oid = {{.Iter}};"},
	}
}

// Setup initializes the database for the benchmark.
func (m *Mysql) Setup() {
	if _, err := m.db.Exec("CREATE DATABASE IF NOT EXISTS dbbench"); err != nil {
		log.Fatalf("failed to create database: %v\n", err)
	}
	if _, err := m.db.Exec("USE dbbench"); err != nil {
		log.Fatalf("failed to USE dbbench: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.simple (id INT PRIMARY KEY, name VARCHAR(10), balance DECIMAL, description VARCHAR(100));"); err != nil {
		log.Fatalf("failed to create table: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.relational_one (oid INT PRIMARY KEY, balance_one DECIMAL);"); err != nil {
		log.Fatalf("failed to create table relational_one: %v\n", err)
	}
	if _, err := m.db.Exec("CREATE TABLE IF NOT EXISTS dbbench.relational_two (balance_two DECIMAL, relation INT PRIMARY KEY, FOREIGN KEY(relation) REFERENCES relational_one(oid));"); err != nil {
		log.Fatalf("failed to create table relational_two: %v\n", err)
	}
	if _, err := m.db.Exec("TRUNCATE dbbench.simple;"); err != nil {
		log.Fatalf("failed to truncate table: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (m *Mysql) Cleanup() {
	if _, err := m.db.Exec("DROP DATABASE dbbench"); err != nil {
		log.Printf("failed drop schema: %v\n", err)
	}
	if err := m.db.Close(); err != nil {
		log.Printf("failed to close connection: %v", err)
	}
}

// Exec executes the given statement on the database.
func (m *Mysql) Exec(stmt string) {
	_, err := m.db.Exec(stmt)
	if err != nil {
		log.Printf("%v failed: %v", stmt, err)
	}
}
