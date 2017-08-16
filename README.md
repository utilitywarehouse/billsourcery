# billsourcery

Perform various operations on the Bill source code, mostly for analysical purposes.

## Requirements

* [Go](https://golang.org/)
* [Mercurial](https://www.mercurial-scm.org/)

## Installation

    go get github.com/utilitywarehouse/billsourcery

## Usage

    billsourcery --help

## Neo4j graph database

### Requirements

* [Java](https://www.java.com)
* [cyphertool](https://github.com/utilitywarehouse/cyphertool)

### Installation

Download and install [Neo4j Community Edition](https://neo4j.com/download/community-edition/)

    $ cd && curl -O https://neo4j.com/artifact.php?name=neo4j-community-3.2.3-unix.tar.gz
    $ tar xvzf neo4j-community-3.2.3-unix.tar.gz
    $ ln -s neo4j-community-3.2.3 neo4j
    $ export PATH=$HOME/neo4j/bin:$PATH

Disable authentication

    $ cp $HOME/neo4j/conf/neo4j.conf{,.orig}
    $ sed -i 's/#dbms.security.auth_enabled=false/dbms.security.auth_enabled=false/' $HOME/neo4j/conf/neo4j.conf

Start neo4j

    $ neo4j start

### Load Bill data

    $ cd && git clone git@github.com:utilitywarehouse/uw-bill-source-history.git
    $ billsourcery --source-root=$HOME/uw-bill-source-history calls-neo | cyphertool run

### Visualise graph data (example queries)

Navigate to [http://localhost:7474/](http://localhost:7474/)

Find called methods that are missing

    MATCH (n:Node) where n.name is null return n

Find examples containing the name "customer"

    MATCH (n:Node) where n.name CONTAINS 'customer' return n

Find methods that are not called from anywhere

    MATCH (m:Method) WHERE NOT (m)<-[:calls]-() RETURN m.name order by m.name

### Progress report

Reset repository and pull latest changes (if required)

    $ cd /path/to/uw-bill-source-history
    $ git reset --hard HEAD && git pull

Strip comments from source files

    $ billsourcery --source-root=$(pwd) strip-comments

Save current status in report

    $ git diff --stat | awk '{print $1 "," $3}' | egrep -e '^(F|P|R|I|M|E).*[0-9]$' > /tmp/report-$(date +%Y%m%d).csv

List top 20 files by number of comments (top files to work on)

    $ git diff --stat | sort -k3 -n -r | head -20

