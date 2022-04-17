---
marp: true
theme: default
paginate: true
size: 16:9

style: |
  section::after {
    /* Layout of pagination content */
    font-size: medium
    }   
    section.custom1 h1 {
       text-align: left;
       padding-top:60px;
       margin-bottom:-20px;
    }
    section.lead h1 {
       text-align: center;
    }
    section.custom ul {
       font-size: 22px;
    }
---
<!-- _class: custom1 -->
<!-- footer: this slide deck is also available as [PDF](https://github.com/RomanBoegli/godbbench/raw/main/docs/slides.pdf)--->
<style scoped>
table { font-size:0.8em;  min-width: 100%; }
th { display:none;}
</style>

# Automated Database Benchmarking Tool
###### Performance Analysis of MySQL, PostgreSQL and Neo4j using Different Data Scenarios
###

.     | .
:-----|:------
**Institute** | Eastern Switzerland University of Applied Science
**Program** | MSE Computer Science
**Module** | DB Seminar
**Author** | Roman Bögli
**Supervisor** | Prof. Stefan F. Keller
**Date** | 14. April 2022
**Context** | Final Presentation



![bg](./assets/OST.svg)

---

# Content

- Relational DBMS vs. Graph-Based DBMS
- Tool `godbbench`
- Synthetic Script & Substitution
- Custom Scripts (`merchant`, `employees`)
- Automation
- Result Analysis
- Open Work

----


# Relational DBMS

- Tables are entities
- Relationships using keys
- Homogenous data 
through schema
- Ideal for **predefinable** & **rigid** 
data use cases

![bg fit right:50% drop-shadow](./assets/merchanterd.drawio.svg)


----


# Graph-Based DBMS

- Attributed nodes and edges
- Relationships are 
first-class citizen
- Heterogenous data 
(schema-less)
- Ideal for **alternating** &  
**highly connected** 
data use cases

![bg 105% right:54% drop-shadow](./assets/friendsgraph.svg)


----


# Query Languages

Query adult customers
```sql
-- SQL
SELECT * FROM Customer c WHERE c.Age >= 18

-- Cyper
MATCH (c:Customer) WHERE c.Age > 18 RETURN c;
```


Show top clients based on revenue
```sql
-- SQL
SELECT c.CustomerId, c.Name, SUM(p.Total) FROM Customer c 
INNER JOIN Purchase p on c.CustomerId = p.CustomerId 
GROUP BY c.CustomerId, c.Name ORDER BY SUM(p.Total) DESC

-- Cyper
MATCH (c:Customer)-[:MAKES]->(p:Purchase)
RETURN c.Name, SUM(p.Total) AS TotalOrderValue ORDER BY TotalOrderValue DESC
```


----

# System Setup

- Requirements:
  - [Docker](https://docs.docker.com/get-docker/)
  - [Go](https://go.dev/doc/install)
  - [godbbench](https://github.com/RomanBoegli/godbbench)

![bg fit right:65% drop-shadow](./assets/systemlandscape.drawio.svg)


----

# Command Line Interface (CLI)

- Open terminal and navigate to the location of `godbbench.go`
`$ cd ~/path/to/godbbench/cmd`

- Interact with `go run godbbench.go` to see flags

![drop-shadow](./assets/gorunmaingoh.png)

----

# Possilbe CLI Commands

```ps
# run synthetic INSERT and SELECT statements against MySQL, each 100x
$ go run godbbench.go mysql --host 127.0.0.1 --port 3306 --user "root" \
        --pass "password" --iter 100 --run "inserts selects"
```

```ps
# run statemets of custom script against Postgres, save results in file
$ go run godbbench.go postgres --host 127.0.0.1 --port 5432 --user "postgres" \
        --pass "password" --iter 100 --script "./path/to/mysql.sql" \
        --writecsv "./path/to/results/mysql.csv"
```

```ps
# merge serveral result files
$ go run godbbench.go mergecsv \
        --rootDir "~/path/with/csv-files/to-be-merged"
        --targetFile "~/anypath/allresults.csv"
```

```ps
# visualize the benchmarking results
$ go run godbbench.go createcharts \
        --dataFile "~/anypath/allresults.csv" --charttype "line"
```

----

# Statement Substitutions

```SQL
INSERT INTO Customer (Id, Name, Birthday) 
VALUES ( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandDate }}');
```

Following expressions will be substituted before the statement is executed:

`{{.Iter}}` --> The iteration counter. Will return 1 when `\benchmark once`.
`{{call .RandIntBetween 1 100}}` --> Random integer between `1` and `100`.
`{{call .RandFloatBetween 0 1}}` --> Random float between `0` and `1`.
`{{call .RandString 3 15}}` --> Random string with length between `3` and `15`.
`{{call .RandDate}}` --> Random date.

----

# Custom Script (`merchant`)

```SQL
-- INIT (illustration purposes)
\benchmark once \name initialize
DROP SCHEMA IF EXISTS godbbench CASCADE; CREATE SCHEMA godbbench;
CREATE TABLE godbbench.order (OrderId INT PRIMARY KEY, CustomerId INT NOT NULL, ... );

-- INSERTS (illustration purposes)
\benchmark loop 1.0 \name inserts
INSERT INTO godbbench.Order (OrderId, CustomerId, CreationDate, Comment) 
VALUES( {{.Iter}}, (SELECT CustomerId FROM godbbench.Customer ORDER BY RANDOM() LIMIT 1), 
        '{{call .RandDate }}', '{{call .RandString 0 50 }}');

-- SELECTS
\benchmark loop 1.0 \name select_simple
SELECT * FROM godbbench.Customer WHERE CustomerId = {{.Iter}} 

\benchmark loop 1.0 \name select_medium
SELECT * FROM godbbench.Product p JOIN godbbench.Supplier s ON ...
  
\benchmark loop 1.0 \name select_complex
SELECT c.CustomerId, c.Name, SUM(li.Quantity * p.UnitSize * p.PricePerUnit) as  ...

-- CLEAN (illustration purposes)
\benchmark once \name clean
DROP SCHEMA IF EXISTS godbbench CASCADE;
```


----

![bg fit drop-shadow](./assets/merchanterd.drawio.svg)

----


![drop-shadow](./assets/merchandneo4j.png)

**Attention:** 
Relational data schemas should not directly be mapped into a graph-world. 
Relationships in graph-based DBs are first-class citizen that can hold information by itself.
 

----


# Custom Script (`employees`)

Show all subordinates of an employee (tree queries)

```SQL
-- use WITH RECURISON notation in Postgres (similar in MySQL)
WITH RECURSIVE hierarchy AS (
    SELECT employeeId, firstname, boss_id, 0 AS level 
    FROM employee 
    WHERE employeeId = {{.Iter}}
  UNION ALL 
    SELECT e.employeeId, e.firstname, e.boss_id, hierarchy.level + 1 AS level 
    FROM employee e JOIN hierarchy ON e.boss_id = hierarchy.employeeId 
) SELECT * FROM hierarchy;
INSERT INTO employee (firstname, boss_id, salary) VALUES ('BigBoss', null, 999999);

-- simpler query using Cypher
MATCH (boss)-[:BOSS_OF*1..]->(sub) WHERE boss.employeeId={{.Iter}} RETURN sub;
```
####
> see example graph on next slide ...

----

![bg 95% drop-shadow](./assets/employeesgraph.svg)

----
<!-- backgroundColor: #60AAF2 -->

# Custom Script Idea: `friends`

Show the shortest acquaintance path of two people (cyclic graph queries)


----

<!-- backgroundColor: white -->

# Automation

```ps
$ bash bashscript.sh
```
![drop-shadow width:700px](./assets/bashscript.png)

>see demo on next slide...

----

![bg 95% drop-shadow](./assets/showcase.gif)

----

# Result Analysis
<!-- footer: 1 second (s) = 1'000'0000 microseconds (μs)  -->
Generating a `chart.html` file to visualize
- average amount of microseconds (`μs`) per benchmark (the lower the better)
- operations per second (the higher the better)
- microseconds per operation (the lower the better) 


![drop-shadow width:700px](./assets/charts.png)

----

![bg 95% ](./assets/chartbar.png)

----

![bg 95% ](./assets/chartline.png)

----

<!-- footer: ""  -->

# Open Work

##

CLI Tool | Custom Scripts | Writing
:--------|:---------------|:-------
✅ Benchmarking<br>✅ Result consolidation<br>✅ Chart generation <br> 🔷 Docker automatization  <br><br>| ✅ `merchant` <br> ✅ `employees`<br>🔷 `friends` <br><br><br>| ✅ Abstract <br>✅ Intro <br>⭕️ System documentation <br>⭕️ Benchmarking approaches <br> ⭕️ Result analysis & conclusion

##

###### Legend: &emsp;✅ done, &emsp; ⭕️ todo, &emsp; 🔷 optional


----
# References
<!-- _class: custom -->

- Chauhan, C., & Kumar, D. (2017). PostgreSQL High Performance Cookbook: Mastering query optimization, database monitoring, and performance-tuning for PostgreSQL. Packt Publishing.
- Codd, E. F. (2002). A Relational Model of Data for Large Shared Data Banks. In M. Broy & E. Denert (Eds.), Software Pioneers (pp. 263–294). Springer Berlin Heidelberg. https://doi.org/10.1007/978-3-642-59412-0_16
- Elmasri, R., & Navathe, S. (2011). Fundamentals of Database Systems (6th ed). Addison-Wesley.
- Gregg, B. (2020). Systems Performance: Enterprise and the Cloud (Second). Addison-Wesley.
- Needham, M., & Hodler, A. E. (2019). Graph Algorithms: Practical Examples in Apache Spark and Neo4j (First edition). O’Reilly Media.
- Peixoto, T. P. (n.d.). What is graph-tool? Graph-Tool. Retrieved 20 March 2022, from https://graph-tool.skewed.de/
- Robinson, I., Webber, J., & Eifrem, E. (2015). Graph Databases: New Opportunities for Connected Data.
- Stopford, B. (2012, August 17). Thinking in Graphs: Neo4J. http://www.benstopford.com/2012/08/17/thinking-in-graphs-neo4j/


----
<!-- _class: lead -->
# Thanks

