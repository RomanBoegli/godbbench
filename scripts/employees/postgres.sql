-- INIT
\benchmark once \name initialize
DROP SCHEMA IF EXISTS godbbench CASCADE; 
CREATE SCHEMA godbbench;
BEGIN;
    CREATE TABLE godbbench.employee (	
        employeeId SERIAL PRIMARY KEY,	
        firstname varchar(50) NOT NULL,	
        boss_id INT NULL, 
        salary INT NULL, 
        FOREIGN KEY (boss_id) REFERENCES godbbench.employee (employeeId));
    INSERT INTO godbbench.employee (firstname, boss_id, salary) VALUES ('BigBoss', null, 999999);
COMMIT;

-- INSERT
\benchmark loop 1.0 \name insert_employee
INSERT INTO godbbench.employee (firstname, boss_id, salary) 
    VALUES ('{{call .RandString 3 10 }}', (SELECT employeeId FROM godbbench.employee ORDER BY RANDOM() LIMIT 1), {{call .RandIntBetween 10000 500000 }});

-- SELECT 1
\benchmark loop 1.0 \name select_before_index
WITH RECURSIVE hierarchy AS (
    SELECT employeeId, firstname, boss_id, 0 AS level 
    FROM godbbench.employee 
    WHERE employeeId = {{.Iter}} 
    UNION ALL 
    SELECT e.employeeId, e.firstname, e.boss_id, hierarchy.level + 1 AS level 
    FROM godbbench.employee e 
    JOIN hierarchy ON e.boss_id = hierarchy.employeeId 
    ) 
SELECT * FROM hierarchy;

-- INDEX
\benchmark once \name create_index
CREATE INDEX index_boss_id ON godbbench.employee USING btree (boss_id);

-- CACHE
\benchmark once \name clear_cache
DISCARD ALL;

-- SELECT 2
\benchmark loop 1.0 \name select_after_index
WITH RECURSIVE hierarchy AS (
    SELECT employeeId, firstname, boss_id, 0 AS level 
    FROM godbbench.employee 
    WHERE employeeId = {{.Iter}} 
    UNION ALL 
    SELECT e.employeeId, e.firstname, e.boss_id, hierarchy.level + 1 AS level 
    FROM godbbench.employee e 
    JOIN hierarchy ON e.boss_id = hierarchy.employeeId 
    ) 
SELECT * FROM hierarchy;

-- CLEAN
\benchmark once \name clean
DROP SCHEMA IF EXISTS godbbench CASCADE;
