\benchmark once \name initialize
MATCH (n) DETACH DELETE n;
CREATE (n:employee {employeeId: 0, firstname: 'BigBoss', salary: 999999 });

// INSERT
\benchmark loop 1.0 \name insert_customer
CREATE (n:employee {employeeId: {{.Iter}}, firstname: '{{call .RandString 3 50 }}', salary: {{call .RandIntBetween 10000 500000 }} }); 
CALL {MATCH (x:employee) WHERE x.employeeId <> {{.Iter}} RETURN x, rand() as randomsort ORDER BY randomsort DESC LIMIT 1} WITH x.employeeId as fk MATCH (a:employee), (b:employee) WHERE a.employeeId = fk AND b.employeeId = {{.Iter}} CREATE (a)-[r:BOSS_OF]->(b);

// SELECT 1
\benchmark loop 1.0 \name select_before_index
MATCH (boss)-[:BOSS_OF*1..]->(sub) WHERE boss.employeeId={{.Iter}} RETURN sub

// INDEX
\benchmark once \name create_index
DROP INDEX ON :employee(employeeId);
CREATE BTREE INDEX index_employeeId FOR (n:employee) ON (n.employeeId);

// CACHE
\benchmark once \name clear_cache
Call db.clearQueryCaches()

// SELECT 1
\benchmark loop 1.0 \name select_after_index
MATCH (boss)-[:BOSS_OF*1..]->(sub) WHERE boss.employeeId={{.Iter}} RETURN sub

// CLEAN
\benchmark once \name clean
MATCH (n) DETACH DELETE n
