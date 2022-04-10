-- INIT
\benchmark once \name initialize
DROP DATABASE IF EXISTS gobench; 
CREATE DATABASE gobench; 
USE gobench;
CREATE TABLE employee (
	employeeId SERIAL PRIMARY KEY,
	firstname varchar(50) NOT NULL,
	boss_id INT NULL REFERENCES employee(employeeId),
  salary INT NULL
);
INSERT INTO employee (firstname, boss_id, salary) VALUES ('BigBoss', null, 999999);

-- INSERT
\benchmark loop 1.0 \name insert_employee
--CREATE TEMPORARY TABLE gobench.tmp SELECT employeeId FROM gobench.employee ORDER BY RAND() LIMIT 1;
SET @fk := (SELECT employeeId FROM gobench.employee ORDER BY RAND() LIMIT 1);
INSERT INTO gobench.employee (firstname, boss_id, salary) VALUES ('{{call .RandString 3 50 }}', @fk, {{call .RandIntBetween 10000 500000 }});
--DROP TEMPORARY TABLE gobench.tmp;

-- SELECT 1
\benchmark loop 1.0 \name select_before_index
WITH RECURSIVE hierarchy AS (
    SELECT employeeId, firstname, boss_id, 0 AS level 
    FROM gobench.employee 
    WHERE employeeId = {{.Iter}} 
    UNION ALL 
    SELECT e.employeeId, e.firstname, e.boss_id, hierarchy.level + 1 AS level 
    FROM gobench.employee e 
    JOIN hierarchy ON e.boss_id = hierarchy.employeeId 
    ) 
SELECT * FROM hierarchy;

-- INDEX
\benchmark once \name create_index
USE gobench;
CREATE INDEX index_boss_id USING BTREE on gobench.employee (boss_id);

-- CACHE
\benchmark once \name clear_cache
FLUSH TABLES;

-- SELECT 2
\benchmark loop 1.0 \name select_after_index
WITH RECURSIVE hierarchy AS (
    SELECT employeeId, firstname, boss_id, 0 AS level 
    FROM gobench.employee 
    WHERE employeeId = {{.Iter}} 
    UNION ALL 
    SELECT e.employeeId, e.firstname, e.boss_id, hierarchy.level + 1 AS level 
    FROM gobench.employee e 
    JOIN hierarchy ON e.boss_id = hierarchy.employeeId 
    ) 
SELECT * FROM hierarchy;

-- CLEAN
\benchmark once \name clean
USE gobench; 
SET FOREIGN_KEY_CHECKS=0; 
DROP DATABASE gobench;
SET FOREIGN_KEY_CHECKS=1;
