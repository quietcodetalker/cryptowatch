package testing

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

type Container struct {
	testcontainers.Container
	URI string
}

type PostgresConfig struct {
	User     string
	Password string
	DBName   string
}

func SetupPostgres(ctx context.Context, cfg PostgresConfig) (*Container, error) {
	port := "5432/tcp"

	dbURL := func(port nat.Port) string {
		return fmt.Sprintf(
			"postgres://%s:%s@localhost:%s/%s?sslmode=disable",
			cfg.User,
			cfg.Password,
			port.Port(),
			cfg.DBName,
		)
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14.2",
		ExposedPorts: []string{port},
		Env: map[string]string{
			"POSTGRES_USER":     cfg.User,
			"POSTGRES_PASSWORD": cfg.Password,
			"POSTGRES_DB":       cfg.DBName,
		},
		WaitingFor: wait.ForSQL(nat.Port(port), "postgres", dbURL).Timeout(time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.User, cfg.Password, ip, mappedPort.Port(), cfg.DBName)

	return &Container{
		Container: container,
		URI:       uri,
	}, nil
}
