package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)


func goDotEnvVariable(key string) string {
    err := godotenv.Load("../.env")
  
    if err != nil {
      log.Fatalf("Error loading .env file")
    }
  
    return os.Getenv(key)
}

func defineDriver(ctx context.Context, dbUri, dbUser, dbPassword string) (neo4j.DriverWithContext, error) {
    driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.BasicAuth(dbUser, dbPassword, ""))
    if err != nil {
        log.Fatal("Error creating Neo4j driver", err)
        return nil, err
    }
    
    err = driver.VerifyConnectivity(ctx)
	if err != nil {
		log.Fatal("Error verifying the connectivity", err)
        return nil, err
	}

    fmt.Println("Connection established")

    return driver, err
}

func createSession(ctx context.Context, driver neo4j.DriverWithContext) neo4j.SessionWithContext {
    return driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

func loadEmployeesAndTheirBoss(ctx context.Context, session neo4j.SessionWithContext) {
    fmt.Printf("Loading employees and their boss...")

    query := `
	    LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-boss.csv' AS row
        CREATE (employee:Employee {name: row.` + "`" + `employee name` + "`" + `})
        MERGE (boss:Employee {name: row.` + "`" + `has boss` + "`" + `})
        WITH employee, boss
        WHERE employee.name <> boss.name
        MERGE (employee)-[:REPORTS_TO]->(boss)
    `

    result, err := session.Run(ctx, query, nil)
    if err != nil {
        log.Fatal("Error running the query", err)
        panic(err)
    }
    
    summary, err := result.Consume(ctx)
    if err != nil {
        log.Fatal("Error consumming the query", err)
        panic(err)
    }
    
    fmt.Printf("Nodes created: %d, Relationships created: %d\n", summary.Counters().NodesCreated(), summary.Counters().RelationshipsCreated())
}

func loadEmployeesAndTheirFriends(ctx context.Context, session neo4j.SessionWithContext) {
    fmt.Printf("Loading employees and their friends...")

    query := `
        LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-friends.csv' AS row
        MERGE (employee:Employee {name: row.employee_name})
        MERGE (friend:Employee {name: row.is_friends_with})
        WITH employee, friend
        WHERE employee.name <> friend.name
        MERGE (employee)-[:FRIENDS_WITH]->(friend)
	`

    result, err := session.Run(ctx, query, nil)
    if err != nil {
        log.Fatal("Error running the query", err)
        panic(err)
    }
    
    summary, err := result.Consume(ctx)
    if err != nil {
        log.Fatal("Error consumming the query", err)
        panic(err)
    }
    
    fmt.Printf("Nodes created: %d, Relationships created: %d\n", summary.Counters().NodesCreated(), summary.Counters().RelationshipsCreated())
}

func loadEmployeesAndTheirSkills(ctx context.Context, session neo4j.SessionWithContext) {
    fmt.Printf("Loading employees and their skills...")

    query := `
        LOAD CSV WITH HEADERS FROM 'https://raw.githubusercontent.com/veben/neo4j_employees/main/resources/data/employees-and-their-skills.csv' AS row
        WITH row, split(row.skills, ",") AS skillList
        UNWIND skillList AS skill
        MERGE (e:Employee {name: row.employee_name})
        MERGE (s:Skill {name: skill})
        MERGE (e)-[:HAS_SKILL]->(s)
	`

    result, err := session.Run(ctx, query, nil)
    if err != nil {
        log.Fatal("Error running the query", err)
        panic(err)
    }
    
    summary, err := result.Consume(ctx)
    if err != nil {
        log.Fatal("Error consumming the query", err)
        panic(err)
    }
    
    fmt.Printf("Nodes created: %d, Relationships created: %d\n", summary.Counters().NodesCreated(), summary.Counters().RelationshipsCreated())
}

func main() {
    ctx := context.Background()

    dbUri := "bolt://localhost:7687"
    dbUser := goDotEnvVariable("DB_USER")
    dbPassword := goDotEnvVariable("DB_PASSWORD")
    
    driver, err := defineDriver(ctx, dbUri, dbUser, dbPassword)
    if err != nil {
        panic(err)
    }
    defer driver.Close(ctx)

	session := createSession(ctx, driver)
	defer session.Close(ctx)

    loadEmployeesAndTheirBoss(ctx, session)
    loadEmployeesAndTheirFriends(ctx, session)
    loadEmployeesAndTheirSkills(ctx, session)
}