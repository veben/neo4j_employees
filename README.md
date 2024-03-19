# Neo4j employees

## I. Installation
- Download Neo4j Desktop: https://neo4j.com/download/neo4j-desktop/?edition=desktop&flavour=winstall64&release=1.5.9&offline=true


## II. Initialization
- Open Neo4j Desktop
- Create project: **Employee Project**
- Create DBMS:
    - Name: Employee DBMS
    - Password: *****
    - Version: 5.18.0
- Click "Start DBMS"
    - It creates databases `system` and `neo4j` by default
- Connect to "Employee DBMS"
    - It creates a powershell execution. Close it crashes the connection

## III. Dataset loading
- Open **Neo4j browser**
- Change csv headers

### 1. Load `employees-and-their-boss.csv` file
- Create "employee "`Employee` nodes with first column
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-boss.csv'
AS row
CREATE (:Employee {name: row.employee_name})
```
> Added 155 labels, created 155 nodes, set 155 properties, completed after 79 ms.

- Create "boss "`Employee` nodes with second column
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-boss.csv'
AS row
MERGE (:Employee {name: row.has_boss})
```
> Added 1 label, created 1 node, set 1 property, completed after 89 ms.

- Look at the inserted nodes
```sh
MATCH (n) RETURN n
```
![alt text](resources/images/employees_nodes.png)

- Create the relashionships between employees and their boss
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-boss.csv'
AS row
MATCH (employee:Employee {name: row.employee_name})
MATCH (boss:Employee {name: row.has_boss})
CREATE (employee)-[:REPORTS_TO]->(boss)
```
> Created 155 relationships, completed after 106 ms.

- Look at the result
```sh
MATCH (n) RETURN n
```
![alt text](resources/images/employees_boss_graph.png)

### 1-bis. Load `employees-and-their-boss.csv` file in 1 step (alternative)
> The execution plan for this query contains the Eager operator, which forces all dependent data to be materialized in main memory before proceeding
Using LOAD CSV with a large data set in a query where the execution plan contains the Eager operator could potentially consume a lot of memory and is likely to not perform well. See the Neo4j Manual entry on the Eager operator for more information and hints on how problems could be avoided.

```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-boss.csv'
AS row
MERGE (employee:Employee {name: row.employee_name})
MERGE (boss:Employee {name: row.has_boss})
WITH employee, boss
WHERE employee.name <> boss.name
MERGE (employee)-[:REPORTS_TO]->(boss)
```
> Added 156 labels, created 156 nodes, set 156 properties, created 155 relationships, completed after 261 ms.

### 2. Load `employees-and-their-friends.csv` file
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-friends.csv'
AS row
MERGE (employee:Employee {name: row.employee_name})
MERGE (friend:Employee {name: row.is_friends_with})
WITH employee, friend
WHERE employee.name <> friend.name
MERGE (employee)-[:FRIENDS_WITH]->(friend)
```
> Created 447 relationships, completed after 185 ms.

- Look at the `FRIENDS_WITH` relationships
```sh
MATCH p=()-[:FRIENDS_WITH]->() RETURN p
```
![alt text](resources/images/friends_with_relationships.png)

### 3. Clean datbase
```sh
MATCH (n)
DETACH DELETE n;
```

## IV. Querying
### Show a hierarchy of all people working under "Darth Vader"
- Show only pairs
```sh
MATCH (boss:Employee {name: "Darth Vader"})<-[:REPORTS_TO*]-(employee)
RETURN boss, employee
```
- Show the complete path
```sh
MATCH (boss:Employee {name: "Darth Vader"})
MATCH path = (boss)<-[:REPORTS_TO*]-(employee)
RETURN path
```

### Show all the people that work on Jacob’s team, but are not friends with Jacob
- Show all employee directed linked to him (team 1)
```sh
MATCH (jacob:Employee {name: "Jacob"})-[:REPORTS_TO]-(employee)
RETURN jacob, employee AS covorker
```
- Show all employee directed linked to him (team 1 bis)
```sh
MATCH (jacob:Employee {name: "Jacob"})
MATCH (employee)-[:REPORTS_TO]->(jacob)
RETURN jacob, employee AS covorker
UNION
MATCH (jacob:Employee {name: "Jacob"})
MATCH (jacob)-[:REPORTS_TO]->(employee)
RETURN jacob, employee AS covorker
```
- Show all employee directed linked to him (team 2)
```sh
MATCH (jacob:Employee {name: "Jacob"})<-[:REPORTS_TO*]-(employee)
RETURN jacob, employee
```
- Show Jacob's friends
```sh
MATCH (jacob:Employee {name: "Jacob"})-[:FRIENDS_WITH]-(friend)
RETURN jacob, friend
```
- Show Jacob's friends 2
```sh
MATCH (jacob:Employee {name: "Jacob"})
MATCH (friend:Employee)
WHERE friend.name <> "Jacob" AND (friend)-[:FRIENDS_WITH]-(jacob:Employee {name: "Jacob"})
RETURN jacob, friend
```
- Show employees that are not friend with Jacob
```sh
MATCH (jacob:Employee {name: "Jacob"})
MATCH (friend:Employee)
WHERE friend.name <> "Jacob" AND NOT (friend)-[:FRIENDS_WITH]-(jacob:Employee {name: "Jacob"})
RETURN jacob, friend
```
- Final query
```sh
MATCH (jacob:Employee {name: "Jacob"})<-[:REPORTS_TO*]-(employee)
WHERE NOT (jacob)<-[:FRIENDS_WITH]-(employee)
RETURN jacob, employee
```

## V. Connect with Tableau
- Download Tableau desktop: https://www.tableau.com/fr-fr/support/releases/desktop/2024.1#esdalt
- Install it
- Download Neo4j **ODBC connector**: https://neo4j.com/bi-connector/
- Install the connector
- Open **ODBC Data Sources** program
- Go to the "System DSN" tab and click "Add".
- Select the Neo4j ODBC Driver from the list of available drivers.
- Configure the driver settings, including the server address (localhost), port number (default is usually 7687), and authentication details (username/password).
- Test the connection
> This leads to this error
```t
[Simba][Neo4j] (22) An error has been thrown from the Neo4j client: 'could not run query: Neo4jError: Neo.ClientError.Statement.SyntaxError (Unknown function 'apoc.version' (line 1, column 8 (offset: 7))
"RETURN apoc.version()"
        ^)'
```
- Install APOC plugin
- Restart Neo4j DBMS
- Test agin the connection
> It works
![alt text](resources/images/odbc_db_connection.png)
- Open Tableau desktop
- Go to the "Data" menu and select "Connect to Data"
- Choose "Other Databases (ODBC)" from the list of available connectors
- Select the Neo4j ODBC DSN you configured earlier
- Fill the connection details
![alt text](resources/images/tableau_db_connection.png)

TODO: Create chart

## VI. Play with more data
### 1. Loading of **skills**
Here is a new dataset adding skill list to the employees:
```
employee_name,skills
Bradley,"Python,Java,C++,SQL"
Meagan,"Java,HTML,CSS,Javascript"
Wayne,"C++,Python,R,SQL"
Annie,"Rust"
Sylvester,"C#,Go"
Ferrari,"Cypher,SQL"
Gavin,"Java,Fortan,Cobol"
Diane,"CSS,HTML,Javascript"
Morgan,"Java"
Mindy,"Go,Java"
Clyde,"C#,C,C++"
Clyde,"Java"
Thad,
```
> This allow to test some behaviour of the loading:
> - multiple values for `skills` column
> - Clyde is present 2 times with different skills
> - Thad does not have skill
- Loading in one time:
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-skills.csv' AS row
WITH row, split(row.skills, ",") AS skillList
UNWIND skillList AS skill
MERGE (e:Employee {name: row.employee_name})
MERGE (s:Skill {name: skill})
MERGE (e)-[:HAS_SKILL]->(s)
```
> Particularity: Thad which has no skills is not loaded
- Loading of employees, then skills, then edges
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-skills.csv' AS row
MERGE (:Employee {name: row.employee_name})
```
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-skills.csv' as row
WITH row, split(row.skills, ",") AS skillList
UNWIND skillList AS skill
MERGE (:Skill {name: skill})
```
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-skills.csv' AS row
WITH row, split(row.skills, ",") AS skillList
UNWIND skillList AS skill
MATCH (e:Employee {name: row.employee_name})
MATCH (s:Skill {name: skill})
MERGE (e)-[:HAS_SKILL]->(s)
```
> Particularity: Thad which has no skills is loaded

### 2. Some queries
- I want to know everything about **Clyde**
```sh
MATCH (e:Employee {name: "Bradley"})-[:REPORTS_TO]->(boss)
OPTIONAL MATCH (e)-[:FRIENDS_WITH]-(friend)
OPTIONAL MATCH (e)-[:HAS_SKILL]->(skill)
RETURN e.name AS Employee,
       boss.name AS Boss,
       collect(DISTINCT friend.name) AS Friends,
       collect(DISTINCT skill.name) AS Skills
```
TODO: Other complex query

## V. Load by script
TODO: Load data with Go script