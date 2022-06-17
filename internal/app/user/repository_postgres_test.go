//go:build integration
// +build integration

package user_test

import (
	"context"
	"cryptowatch/internal/app/user"
	"cryptowatch/pkg/config"
	"cryptowatch/pkg/util"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"
)

type PostgresRepoTestSuite struct {
	db   *pgxpool.Pool
	repo user.Repository

	suite.Suite
}

func (s *PostgresRepoTestSuite) SetupSuite() {
	_, filename, _, _ := runtime.Caller(0)
	rootDir := path.Join(path.Dir(filename), "../../..")

	cfg, err := config.LoadConfig(
		path.Join(rootDir, "configs"),
		"test",
	)
	require.NoError(s.T(), err)

	dbSource := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	s.db, err = util.OpenDB(dbSource)
	require.NoError(s.T(), err)

	s.repo = user.NewPostgresRepo(s.db)
}

func (s *PostgresRepoTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *PostgresRepoTestSuite) TearDownTest() {
	query := `
	SELECT tablename
	FROM pg_catalog.pg_tables
	WHERE schemaname != 'pg_catalog' AND
				schemaname != 'information_schema' AND
				tablename != 'schema_migrations'
	`
	rows, err := s.db.Query(context.Background(), query)
	require.NoError(s.T(), err)
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		require.NoError(s.T(), err)

		truncateQuery := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)
		_, err = s.db.Exec(context.Background(), truncateQuery)
		require.NoError(s.T(), err)
	}

	rows.Close()
	require.NoError(s.T(), rows.Err())
}

func (s *PostgresRepoTestSuite) TestCreate() {
	now := time.Now()
	reqs := []user.RepoCreateReq{
		{
			Username:     "username2",
			PasswordHash: "password2",
			FirstName:    "firstname2",
			LastName:     "lastname2",
		},
	}
	users := s.seedUsers(reqs)

	user1 := user.User{
		Username:     "username1",
		PasswordHash: "password1",
		FirstName:    "firstname1",
		LastName:     "lastname1",
	}
	user2 := users[0]

	tests := []struct {
		name        string
		req         user.RepoCreateReq
		expectedRes *user.User
		err         error
	}{
		{
			name: "OK",
			req: user.RepoCreateReq{
				Username:     user1.Username,
				PasswordHash: user1.PasswordHash,
				FirstName:    user1.FirstName,
				LastName:     user1.LastName,
			},
			expectedRes: &user.User{
				Username:     user1.Username,
				PasswordHash: user1.PasswordHash,
				FirstName:    user1.FirstName,
				LastName:     user1.LastName,
				CreateTime:   now,
			},
			err: nil,
		},
		{
			name: "Too short username",
			req: user.RepoCreateReq{
				Username:     strings.Repeat("a", 3),
				PasswordHash: user1.PasswordHash,
				FirstName:    user1.FirstName,
				LastName:     user1.LastName,
			},
			expectedRes: nil,
			err:         user.ErrInvalidArgument,
		},
		{
			name: "Too long username",
			req: user.RepoCreateReq{
				Username:     strings.Repeat("a", 17),
				PasswordHash: user1.PasswordHash,
				FirstName:    user1.FirstName,
				LastName:     user1.LastName,
			},
			expectedRes: nil,
			err:         user.ErrInvalidArgument,
		},
		{
			name: "Username duplicate",
			req: user.RepoCreateReq{
				Username:     user2.Username,
				PasswordHash: user2.PasswordHash,
				FirstName:    user2.FirstName,
				LastName:     user2.LastName,
			},
			expectedRes: nil,
			err:         user.ErrFailedPrecondition,
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			res, err := s.repo.Create(context.Background(), tt.req)
			assert.ErrorIs(s.T(), err, tt.err)
			if tt.expectedRes == nil {
				assert.Equal(s.T(), tt.expectedRes, res)
			} else {
				require.NotEmpty(s.T(), res)

				assert.Equal(s.T(), tt.expectedRes.Username, res.Username)
				assert.Equal(s.T(), tt.expectedRes.PasswordHash, res.PasswordHash)
				assert.Equal(s.T(), tt.expectedRes.FirstName, res.FirstName)
				assert.Equal(s.T(), tt.expectedRes.LastName, res.LastName)

				assert.WithinDuration(s.T(), now, res.CreateTime, time.Second)
			}
		})
	}
}

func (s *PostgresRepoTestSuite) TestGetByUsername() {
	now := time.Now()

	reqs := []user.RepoCreateReq{
		{
			Username:     "username2",
			PasswordHash: "password2",
			FirstName:    "firstname2",
			LastName:     "lastname2",
		},
	}
	users := s.seedUsers(reqs)

	tests := []struct {
		name        string
		username    string
		expectedRes *user.User
		err         error
	}{
		{
			name:        "Not found",
			username:    "username1",
			expectedRes: nil,
			err:         user.ErrNotFound,
		},
		{
			name:        "OK",
			username:    users[0].Username,
			expectedRes: users[0],
			err:         nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			res, err := s.repo.GetByUsername(context.Background(), tt.username)
			assert.ErrorIs(s.T(), err, tt.err)
			if tt.expectedRes == nil {
				assert.Equal(s.T(), tt.expectedRes, res)
			} else {
				require.NotEmpty(s.T(), res)

				assert.Equal(s.T(), tt.expectedRes.ID, res.ID)
				assert.Equal(s.T(), tt.expectedRes.Username, res.Username)
				assert.Equal(s.T(), tt.expectedRes.PasswordHash, res.PasswordHash)
				assert.Equal(s.T(), tt.expectedRes.FirstName, res.FirstName)
				assert.Equal(s.T(), tt.expectedRes.LastName, res.LastName)

				assert.WithinDuration(s.T(), now, res.CreateTime, time.Second)
			}
		})
	}
}

func TestRepositoryPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresRepoTestSuite))
}
