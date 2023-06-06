package initializers

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"testing"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// Create a new container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_DB":       "go-kafka-rest-test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	// Start the container
	pg, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	// Get the host
	host, err := pg.Host(context.Background())
	if err != nil {
		panic(err)
	}

	// Get the port
	port, err := pg.MappedPort(context.Background(), "5432")
	if err != nil {
		panic(err)
	}

	// Connect to the database
	db, err = sql.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@%s:%s/go-kafka-rest-test?sslmode=disable", host, port.Port()))
	if err != nil {
		panic(err)
	}

	// Run the tests
	code := m.Run()

	// Stop the container
	if err := pg.Terminate(context.Background()); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func TestConnectDB(t *testing.T) {
	// Check that DB is not nil after connection
	assert.NotNil(t, db)

	// Perform any other checks to validate successful connection
	// You may want to execute some queries and check their result
}

func TestGetDB(t *testing.T) {
	// Assuming ConnectDB has been successfully called before
	db := GetDB()

	// Check that GetDB returns the DB instance
	assert.Equal(t, DB, db)

	// Perform any other checks to validate that the DB instance works
	// You may want to execute some queries and check their result
}
