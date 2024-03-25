# Neo4j employees

## I. Definition of Neo4j DBMS
> ⚠️ This required **Docker** and **docker-compose** installed on your system
- Define a `docker-compose.yml` file with the following content
```yml
services:
  neo4j: 
    image: neo4j:latest
    hostname: neo4j
    container_name: neo4j
    ports:
      - "7474:7474" # port mapping for browser interface
      - "7687:7687" # port mapping for Bolt protocol
    volumes:
      - ./neo4j/data:/data
      - ./neo4j/logs:/logs
    environment:
      - NEO4J_AUTH=$DB_USER/$DB_PASSWORD
      - NEO4J_PLUGINS=["apoc"] # plugin needed to connect with Tableau

volumes:
  neo4j:
```
- Create a `.env` file to initialize `$DB_USER` `$DB_PASSWORD` environment variables. This file has also to be put in the `.gitignore` to not be pushed on remote repository.
```text
DB_USER=****
DB_PASSWORD=*****
```

## II. Initialization of Neo4j DBMS
- Initialize the DBMS with **docker-compose**
```sh
docker-compose up
```
- Open your the browser interface with this URL: `http://localhost:7474/browser/`
- Connect to the Neo4j DBMS
> - We used `neo4j:latest` docker image - For now the neo4j version is `5.18.1`
> - Databases `system` and `neo4j` are created by default
- We can use the **Neo4j shell** to see server status using this command:
```sh
:server status
```

## III. Dataset loading
- Open **Neo4j browser**

### 1. Load `employees-and-their-boss.csv` file
- Create "employee "`Employee` nodes with first column
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-boss.csv' AS row
CREATE (:Employee {name: row.`employee name`})
```
> 💡 Added 155 labels, created 155 nodes, set 155 properties, completed after 79 ms.

- Create "boss "`Employee` nodes with second column
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-boss.csv' AS row
MERGE (:Employee {name: row.`has boss`})
```
> 💡 Added 1 label, created 1 node, set 1 property, completed after 89 ms.
> I used `MERGE` and not `CREATE`. Otherwise it would create duplicate employees

- Look at the inserted nodes
```sh
MATCH (n:Employee) RETURN n
```
<img src="resources/images/employees_nodes.png" alt="employees_nodes.png" style="width:500px;height:auto;">

- Create the relashionships between employees and their boss
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-boss.csv' AS row
MATCH (employee:Employee {name: row.`employee name`})
MATCH (boss:Employee {name: row.`has boss`})
WITH employee, boss
WHERE employee.name <> boss.name
CREATE (employee)-[:REPORTS_TO]->(boss)
```
> 💡 Created 155 relationships, completed after 106 ms.

- Look at the result
```sh
MATCH (n) RETURN n
```
<img src="resources/images/employees_boss_graph.png" alt="employees_boss_graph.png" style="width:500px;height:auto;">

- See who is the employee that reports to no one:
```sh
MATCH (employee:Employee)
WHERE not (employee)-[:REPORTS_TO]->(:Employee)
RETURN employee
```

### 1-bis. Load `employees-and-their-boss.csv` file in 1 step (alternative)
- Let's try to do the 3 steps in 1
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-boss.csv' AS row
CREATE (employee:Employee {name: row.`employee name`})
MERGE (boss:Employee {name: row.`has boss`})
WITH employee, boss
WHERE employee.name <> boss.name
MERGE (employee)-[:REPORTS_TO]->(boss)
```

- The following message is displayed:
> The execution plan for this query contains the Eager operator, which forces all dependent data to be materialized in main memory before proceeding
Using LOAD CSV with a large data set in a query where the execution plan contains the Eager operator could potentially consume a lot of memory and is likely to not perform well. See the Neo4j Manual entry on the Eager operator for more information and hints on how problems could be avoided.
- We can try to dive in the execution of the query using `EXPLAIN` or `PROFILE` keywords on top of the query
- If we validate the query:
> 💡 Added 156 labels, created 156 nodes, set 156 properties, created 155 relationships, completed after 261 ms.

### 2. Load `employees-and-their-friends.csv` file
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-friends.csv' AS row
MERGE (employee:Employee {name: row.employee_name})
MERGE (friend:Employee {name: row.is_friends_with})
WITH employee, friend
WHERE employee.name <> friend.name
MERGE (employee)-[:FRIENDS_WITH]->(friend)
```
> 💡 Created 447 relationships, completed after 185 ms.

- Look at employees friend with themselves
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-friends.csv' AS row
MATCH (employee:Employee {name: row.employee_name})
MATCH (friend:Employee {name: row.is_friends_with})
WITH employee, friend
WHERE employee.name = friend.name
RETURN employee.name, friend.name
```

- Look at the `FRIENDS_WITH` relationships
```sh
MATCH p=()-[:FRIENDS_WITH]->()
RETURN p
```
<img src="resources/images/friends_with_relationships.png" alt="friends_with_relationships.png" style="width:500px;height:auto;">

- If you want to display a table with employees and friends
```sh
MATCH (employee)-[:FRIENDS_WITH]->(friend)
RETURN employee.name, friend.name
```

### 3. Clean database (if needed)
- Delete all relashionships and node
```sh
MATCH (n)-[r]-() DELETE r
MATCH (n) delete n
```
- In one time
```sh
MATCH (n)
DETACH DELETE n;
```

## IV. Querying
### 1. Show a hierarchy of all people working under "Darth Vader"
```sh
MATCH (boss:Employee {name: "Darth Vader"})<-[:REPORTS_TO*]-(employee:Employee)
RETURN boss, employee
```

### 2. Show all the people that work on Jacob’s team, but are not friends with Jacob
- Show all employee directed linked to him (team 1)
```sh
MATCH (jacob:Employee {name: "Jacob"})-[:REPORTS_TO]-(employee:Employee)
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
- Show people working directly under him (team 2)
```sh
MATCH (employee)-[:REPORTS_TO]->(jacob:Employee {name: "Jacob"})
RETURN jacob, employee
```
- Show Jacob's friends
```sh
MATCH (friend:Employee)-[:FRIENDS_WITH]->(jacob:Employee {name: "Jacob"})
RETURN jacob, friend
```
- Show employees that are not friend with Jacob
```sh
MATCH (employee:Employee)
MATCH (jacob:Employee {name: "Jacob"})
WHERE NOT (employee)-[:FRIENDS_WITH]->(jacob) AND employee.name <> jacob.name
RETURN jacob.name, employee.name
```
- Final query
```sh
MATCH (employee:Employee)-[:REPORTS_TO]->(jacob:Employee {name: 'Jacob'})
WHERE NOT (employee)-[:FRIENDS_WITH]->(jacob)
RETURN jacob, employee
```

## V. Play with more data
### 1. Loading of **skills**
Here is a new dataset adding skill list to the employees:
```text
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
Steve,"Cypher,Go,Kotlin,Java"
Wilson,"C#,Go,Java"
```
> This allow to test some behaviours of the loading:
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
- Show the graph:
```sh
MATCH p=()-[r:HAS_SKILL]->() RETURN p
```
> 📢 Particularity: Thad which has no skills is not loaded
- Loading of employees, then skills, then edges
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-skills.csv' AS row
MERGE (:Employee {name: row.employee_name})
```
```sh
LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-skills.csv' AS row
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
> 📢 Particularity: Thad which has no skills is loaded

### 2. Some more queries
- I want to know everything about **Bradley**
```sh
MATCH (brad:Employee {name: "Bradley"})-[:REPORTS_TO]->(boss)
OPTIONAL MATCH (brad)-[:FRIENDS_WITH]-(friend)
OPTIONAL MATCH (brad)-[:HAS_SKILL]->(skill)
RETURN brad.name AS Employee,
       boss.name AS Boss,
       collect(DISTINCT friend.name) AS Friends,
       collect(DISTINCT skill.name) AS Skills
```

- Employees who have the same boss and are friends
```sh
MATCH (employee:Employee)-[:REPORTS_TO]->(boss:Employee)
MATCH (employee)-[:FRIENDS_WITH]->(friend:Employee)
MATCH (friend)-[:REPORTS_TO]->(boss)
RETURN employee.name, friend.name, boss.name
```
- Same with only one `MATCH`
```sh
MATCH (employee:Employee)-[:REPORTS_TO]->(boss:Employee)<-[:REPORTS_TO]-(friend:Employee)<-[:FRIENDS_WITH]-(employee)
RETURN employee.name, friend.name, boss.name
```
- Retrieve employees who have skills in common with their friends, and also have the same boss:
```sh
MATCH (employee:Employee)-[:FRIENDS_WITH]->(friend:Employee)
MATCH (employee)-[:HAS_SKILL]->(skill:Skill)<-[:HAS_SKILL]-(friend)
MATCH (employee)-[:REPORTS_TO]->(boss:Employee)<-[:REPORTS_TO]-(friend)
RETURN employee.name, friend.name, boss.name, COLLECT(skill.name)
```

## VI. Load by script
> ⚠️ This required **Golang** installed on your system
- Go to `src` folder
- Download the needed libraries:
```sh
go mod tidy
```
- Launch the `insert_data.go` script:
```sh
go run insert_data.go
```

## VII. Connect with Tableau
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
- Test again the connection
> It works

<img src="resources/images/odbc_db_connection.png" alt="odbc_db_connection.png" style="width:500px;height:auto;">

- Open Tableau desktop
- Go to the "Data" menu and select "Connect to Data"
- Choose "Other Databases (ODBC)" from the list of available connectors
- Select the Neo4j ODBC DSN you configured earlier
- Fill the connection details

<img src="resources/images/tableau_db_connection.png" alt="tableau_db_connection.png" style="width:500px;height:auto;">

⌛ TODO: Create chart

## VIII. Cleaning
- Go to `neo4j_employees` folder
- Destroy docker container and associated components:
```sh
docker-compose down
```