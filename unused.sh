#!/bin/bash
echo "MATCH (n:Table) where not exists { match (n:Table)<-[r:references]-(m) where NOT m:Index and NOT m:Field AND r IS NOT NULL} RETURN distinct n.name as table_name" | cypher-shell  > unused_tables.csv
echo "MATCH (i:Index)-[:references]->(t:Table) WHERE NOT (i)<-[:references]-() RETURN t.name as table_name, i.name as index_name order by table_name, index_name" | cypher-shell > unused_indexes.csv
echo "MATCH (pp:PublicProcedure) WHERE NOT (pp)<-[:references]-() and not pp:Used RETURN pp.name as public_procedure_name order by public_procedure_name" | cypher-shell > unused_pp.txt
