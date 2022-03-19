# Abstract
The goal of this paper is to analyse the differences of relational and graph-based database management systems. The representatives used as concrete implementation of these two paradigms will be PostgreSQL (relational) and Neo4j (graph-based).

The first part of this work will elaborate the background of these technologies with focus on the history, popular use cases, as well as (dis)advantages. Furthermore, the key differences will be outlined in the applicable query languages, namely SQL and Cypher.

The main part is dedicated to a setup and execution of a benchmarking test. The goal is to measure and compare the performances of standard database statements used to create, read, update, and delete data. Therefore, a test console application was developed using *Go* in order to consistently and automatically test the given statements with database instances running in *Docker*.

Finally, the benchmarking results are consolidated and interpreted. The findings will be discussed alongside with concrete recommendations in order to facilitate future decision on the given database paradigm.

# Relational Database Systems `20%`
- History
- Use Cases
- (Dis)Advantages

# Graph-Based Database Systems `20%`
- History
- Use Cases
- (Dis)Advantages

# Query Languages `20%`
- General way of working
- Data Definition Language
- Data Manipulation Language

# Benchmark `40%`
- Intro
- Important to Know (e.g. warm up, caching, etc.)

## Strategy and Goals
- Explanation of Automatised Tests 
- Evaluation Criteria (Performance)

## Setup
- Hardware
- Software
- system setup (docker, etc.)
- Sample Data

## Results
- Consolidation
- Interpretation

# Discussion
- Are Graph-Based really always better?

# Acknowledgements
The command line tool was heavily inspired by the previous work of Simon Jürgensmeyer's [dbbench](https://github.com/sj14/dbbench) project.