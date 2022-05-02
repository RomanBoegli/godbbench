-- INIT
\benchmark once \name initialize
DROP DATABASE IF EXISTS godbbench; 
CREATE DATABASE godbbench; 
USE godbbench;
START TRANSACTION;
  CREATE TABLE employee (
    employee_id SERIAL PRIMARY KEY,
    first_name varchar(50) NOT NULL,
    boss_id INT NULL REFERENCES employee(employee_id),
    salary INT NULL
  );
  INSERT INTO employee (first_name, boss_id, salary) VALUES ('BigBoss', null, 999999);
COMMIT;

-- INSERT
\benchmark loop 1.0 \name insert_employee
SET @fk := (SELECT employee_id FROM godbbench.employee ORDER BY RAND() LIMIT 1);
INSERT INTO godbbench.employee (first_name, boss_id, salary) VALUES ('{{call .RandString 3 50 }}', @fk, {{call .RandIntBetween 10000 500000 }});

-- SELECT 1
\benchmark loop 1.0 \name select_before_index
WITH RECURSIVE hierarchy AS (
    SELECT employee_id, first_name, boss_id, 0 AS level 
    FROM godbbench.employee 
    WHERE employee_id = {{.Iter}} 
    UNION ALL 
    SELECT e.employee_id, e.first_name, e.boss_id, hierarchy.level + 1 AS level 
    FROM godbbench.employee e 
    JOIN hierarchy ON e.boss_id = hierarchy.employee_id 
    ) 
SELECT * FROM hierarchy;

-- INDEX
\benchmark once \name create_index
USE godbbench;
CREATE INDEX index_boss_id USING BTREE on godbbench.employee (boss_id);


-- CACHE
\benchmark once \name clear_cache
FLUSH TABLES;

-- SELECT 2
\benchmark loop 1.0 \name select_after_index
WITH RECURSIVE hierarchy AS (
    SELECT employee_id, first_name, boss_id, 0 AS level 
    FROM godbbench.employee 
    WHERE employee_id = {{.Iter}} 
    UNION ALL 
    SELECT e.employee_id, e.first_name, e.boss_id, hierarchy.level + 1 AS level 
    FROM godbbench.employee e 
    JOIN hierarchy ON e.boss_id = hierarchy.employee_id 
    ) 
SELECT * FROM hierarchy;

-- CLEAN
\benchmark once \name clean
USE godbbench; 
SET FOREIGN_KEY_CHECKS=0; 
DROP DATABASE godbbench;
SET FOREIGN_KEY_CHECKS=1;
