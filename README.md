# billsourcery

A tool for analysing the source code of Equinox applications.

## Requirements

* [Go](https://golang.org/)

## Installation

    go install github.com/utilitywarehouse/billsourcery@latest

## Usage

    billsourcery --help

## Neo4j graph database

If using the neo4j output from billsourcery, you may wish to install and use neo4j.

### Neo4J Installation

Download and install [Neo4j](https://neo4j.com/)

For simple local use, disable authentication by setting `dbms.security.auth_enabled=false` in `neo4j.conf`

### Load analysed bill source code data into neo4j

This may take some time to import (e.g. 10 or 20 minutes)

    $ cd && git clone git@github.com:utilitywarehouse/uw-bill-source-history.git
    $ billsourcery --source-root=${PATH_TO_BILL_SOURCE} calls-neo | cypher-shell

### Visualise graph data (example queries)

Navigate to [http://localhost:7474/](http://localhost:7474/)

Find missing nodes that are referenced from elsewhere:

    `MATCH (n:Missing) return n`

Find nodes containing the name "customer":

    `MATCH (n:Node) where n.name CONTAINS 'customer' return n`

Find methods that are not called from anywhere:

    `MATCH (m:Method) WHERE NOT (m)<-[:references]-() RETURN m.name order by m.name`

Find method `nrg_sweep2` and everything it references, recursively.  Exclude fields and work areas for clarity:

    `MATCH p=(n:Node)-[r:references*]->(x:Node) where lower(n.name)="nrg_sweep2" and not(x:Field) and not(x:WorkArea) RETURN p`

Find table `ginv` and everything that references it, recursively.

    `MATCH p=(n:Node)-[r:references*]->(x:Table) where lower(x.name)="ginv" RETURN p`

Find the most referenced tables:

    `MATCH (n)-[references]->(t:Table) RETURN  t.name, count(n) order by count(n) desc limit 20`


