package databases

import (
	"fmt"
	"log"
	"strings"

	"github.com/RomanBoegli/gobench/benchmark"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// neo4j implements the bencher interface.
type Neo4j struct {
	driver neo4j.Driver
}

// NewNeo4J returns a new neo4j bencher.
func NewNeo4J(host string, port int, user, password string) *Neo4j {

	if port == 0 {
		port = 7687
	}

	uri := fmt.Sprintf("neo4j://%v:%v", host, port)
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil
	}

	if err != nil {
		log.Fatalf("failed to create session: %v\n", err)
	}

	p := &Neo4j{driver: driver}
	return p
}

// Benchmarks returns the individual benchmark functions for the cassandra db.
// TODO: update is not like other db statements balance = balance + balance!
func (c *Neo4j) Benchmarks() []benchmark.Benchmark {
	return []benchmark.Benchmark{
		{Name: "inserts", Type: benchmark.TypeLoop, IterRatio: 1.0, Parallel: false, Stmt: "CREATE (ee:Person {id: {{.Iter}}, from: 'Switzerland', balance: {{call .RandInt63}}});"},
		{Name: "selects", Type: benchmark.TypeLoop, IterRatio: 1.0, Parallel: false, Stmt: "MATCH (ee:Person) WHERE ee.id = {{.Iter}} RETURN ee;"},
		{Name: "updates", Type: benchmark.TypeLoop, IterRatio: 1.0, Parallel: false, Stmt: "MATCH (ee:Person {id: {{.Iter}} }) SET ee.balance = {{call .RandInt63}};"},
		{Name: "deletes", Type: benchmark.TypeLoop, IterRatio: 1.0, Parallel: false, Stmt: "MATCH (n:Person {id: {{.Iter}} }) DELETE n"},
	}
}

// Setup initializes the database for the benchmark.
func (c *Neo4j) Setup() {
	session := c.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	_, err := session.Run("MATCH (n) DETACH DELETE n", nil)
	if err != nil {
		log.Fatalf("failed to create keyspace: %v\n", err)
	}
}

// Cleanup removes all remaining benchmarking data.
func (c *Neo4j) Cleanup(closeConnection bool) {
	session := c.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	if _, err := session.Run("MATCH (n) OPTIONAL MATCH (n)-[r]-() RETURN n,r", nil); err != nil {
		log.Printf("failed to drop table: %v\n", err)
	}

	if closeConnection {
		session.Close()
		c.driver.Close()
	}
}

// Exec executes the given statement on the database.
func (c *Neo4j) Exec(stmt string) {
	session := c.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	singleStmts := strings.Split(stmt, ";")
	for _, stmt := range singleStmts {
		if stmt != "" {
			if _, err := session.Run(stmt, nil); err != nil {
				log.Fatalf("%v: failed(!): %v\n", stmt, err)
			}
		}
	}
}

/*
// Exec executes the given statement on the database using transactions.
func (c *Neo4j) Exec(stmt string) {
	session := c.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	transaction, err := session.BeginTransaction()
	if err != nil {
		panic(err)
	}
	defer transaction.Close()
	singleStmts := strings.Split(stmt, ";")
	for _, stmt := range singleStmts {
		if stmt != "" {
			if _, err := transaction.Run(stmt, nil); err != nil {
				log.Fatalf("%v: failed(!): %v\n", stmt, err)
			}
		}
	}

	transaction.Commit()

}
*/
