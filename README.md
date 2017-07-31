billsourcery
============

Perform various operations on the Bill source code, mostly for analysical purposes.

Installation
------------
Assuming you have a working [Go](https://golang.org/) installation:

`go get github.com/utilitywarehouse/billsourcery`

Usage
-----

`billsourcery --help`



Examples of neo4j queries
-------------------------

After using `billsourcery calls-neo` and importing the result into neo4j, here are some example queries that might be useful.

* Find called methods that are missing:
	`MATCH (n:Node) where n.name is null return n`
* Find examples containing the name "customer":
	`MATCH (n:Node) where n.name CONTAINS 'customer' return n;`
* Find methods that are not called from anywhere:
	`MATCH (m:Method) WHERE NOT (m)<-[:calls]-() RETURN m.name order by m.name`
