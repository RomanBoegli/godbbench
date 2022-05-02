#!/bin/bash

# This script assumes the docker containers are already up and running.
# All paths should be indicated relatively to this file.

####################
#### VARIABLES #####
####################

# general
HOST="127.0.0.1"
MULTIPLICITIES=("10" "100" "1000" "10000")
THREADS=15
PATH_TO_CLI="./godbbench.go"
SCRIPT_BASE_PATH="../scripts"
SCRIPT_SET="employees"
RESULT_BASE_PATH="../tmp/results"
CHART_TYPE="line"

# mysql
MYSQL_PORT="3306"
MYSQL_USER="root"
MYSQL_PASS="password"

# neo4j
NEO_PORT="7687"
NEO_USER="neo4j"
NEO_PASS="password"

# postgres
POSTGRES_PORT="5432"
POSTGRES_USER="postgres"
POSTGRES_PASS="password"


########################
#### ACTUAL SCRIPT #####
########################

start_time=`date +%s`
echo -e "\nSTART BENCHMARKING...\n"
for MULT in "${MULTIPLICITIES[@]}"; do
    echo $(for i in $(seq 1 50); do printf "_"; done) 
    echo -e "\nITERATIONS: ${MULT}"
    
    echo -e "\nTEST MYSQL"
    go run $PATH_TO_CLI mysql \
        --host $HOST \
        --port $MYSQL_PORT \
        --user $MYSQL_USER \
        --pass $MYSQL_PASS \
        --iter $MULT \
        --threads $THREADS \
        --script "${SCRIPT_BASE_PATH}/${SCRIPT_SET}/mysql.sql" \
        --writecsv "${RESULT_BASE_PATH}/${SCRIPT_SET}/mysql_${MULT}.csv"

    echo -e "\nTEST NEO4J"
    go run $PATH_TO_CLI neo4j \
        --host $HOST \
        --port $NEO_PORT \
        --user $NEO_USER \
        --pass $NEO_PASS \
        --iter $MULT \
        --threads $THREADS \
        --script "${SCRIPT_BASE_PATH}/${SCRIPT_SET}/neo4j.cql" \
        --writecsv "${RESULT_BASE_PATH}/${SCRIPT_SET}/neo4j_${MULT}.csv"

    echo -e "\nTEST POSTGRES"
    go run $PATH_TO_CLI postgres \
        --host $HOST \
        --port $POSTGRES_PORT \
        --user $POSTGRES_USER \
        --pass $POSTGRES_PASS \
        --iter $MULT \
        --threads $THREADS \
        --script "${SCRIPT_BASE_PATH}/${SCRIPT_SET}/postgres.sql" \
        --writecsv "${RESULT_BASE_PATH}/${SCRIPT_SET}/postgres_${MULT}.csv"
done

echo -e "\n"
echo $(for i in $(seq 1 50); do printf "#"; done) 

echo -e "\nMERGE RESULTS"
go run $PATH_TO_CLI mergecsv \
    --rootDir "${RESULT_BASE_PATH}/${SCRIPT_SET}" \
    --targetFile "${RESULT_BASE_PATH}/${SCRIPT_SET}/merged_results.csv" 

echo -e "\nCREATE CHARTS"
go run $PATH_TO_CLI createcharts \
    --dataFile "${RESULT_BASE_PATH}/${SCRIPT_SET}/merged_results.csv" \
    --chartType $CHART_TYPE

echo $(for i in $(seq 1 50); do printf "#"; done) 
echo -e "\nTOTAL RUN TIME: " $(expr `date +%s` - $start_time) s
