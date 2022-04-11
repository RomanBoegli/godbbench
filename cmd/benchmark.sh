#!/bin/bash

# This script assumes the docker containers are already up and running.

####################
#### VARIABLES #####
####################

script_set="employees"

# general
db_host="127.0.0.1"
MULTIPLICITIES=("10" "100" "1000" "10000")
threads=15
gobench_main_path="/Users/rbo/Documents/Gits/gobench/cmd/main.go"
script_base_path="/Users/rbo/Documents/Gits/gobench/scripts"
result_base_path="/Users/rbo/Documents/Gits/gobench/tmp/results"
chart_type="line"

# mysql
mysql_port="3306"
mysql_user="root"
mysql_pass="password"

# neo4j
neo_port="7687"
neo_user="neo4j"
neo_pass="password"

# postgres
postgres_port="5432"
postgres_user="postgres"
postgres_pass="password"


########################
#### ACTUAL SCRIPT #####
########################

start_time=`date +%s`
echo -e "\nSTART BENCHMARKING...\n"
for MULT in "${MULTIPLICITIES[@]}"; do
    echo $(for i in $(seq 1 50); do printf "_"; done) 
    echo -e "\nITERATIONS: ${MULT}"
    
    echo -e "\nTEST MYSQL"
    go run $gobench_main_path mysql \
        --host $db_host \
        --port $mysql_port \
        --user $mysql_user \
        --pass $mysql_pass \
        --iter $MULT \
        --threads $threads \
        --script "${script_base_path}/${script_set}/mysql.sql" \
        --writecsv "${result_base_path}/${script_set}/mysql_${MULT}.csv"

    echo -e "\nTEST NEO4J"
    go run $gobench_main_path neo4j \
        --host $db_host \
        --port $neo_port \
        --user $neo_user \
        --pass $neo_pass \
        --iter $MULT \
        --threads $threads \
        --script "${script_base_path}/${script_set}/neo4j.cql" \
        --writecsv "${result_base_path}/${script_set}/neo4j_${MULT}.csv"

    echo -e "\nTEST POSTGRES"
    go run $gobench_main_path postgres \
        --host $db_host \
        --port $postgres_port \
        --user $postgres_user \
        --pass $postgres_pass \
        --iter $MULT \
        --threads $threads \
        --script "${script_base_path}/${script_set}/postgres.sql" \
        --writecsv "${result_base_path}/${script_set}/postgres_${MULT}.csv"
done

echo -e "\n"
echo $(for i in $(seq 1 50); do printf "#"; done) 

echo -e "\nMERGE RESULTS"
go run $gobench_main_path mergecsv \
    --rootDir "${result_base_path}/${script_set}" \
    --targetFile "${result_base_path}/${script_set}/merged_results.csv" 

echo -e "\nCREATE CHARTS"
go run $gobench_main_path createcharts \
    --dataFile "${result_base_path}/${script_set}/merged_results.csv" \
    --charttype $chart_type

echo $(for i in $(seq 1 50); do printf "#"; done) 
echo -e "\nTOTAL RUN TIME: " $(expr `date +%s` - $start_time) s
