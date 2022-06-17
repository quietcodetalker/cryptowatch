//go:build integration
// +build integration

package user_test

import (
	"context"
	"cryptowatch/internal/app/user"
	"github.com/stretchr/testify/require"
)

func (s *PostgresRepoTestSuite) seedUsers(reqs []user.RepoCreateReq) []*user.User {
	users := make([]*user.User, 0, len(reqs))

	var u *user.User
	var err error

	for _, r := range reqs {
		u, err = s.repo.Create(context.Background(), r)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), u)
		users = append(users, u)
	}

	return users
}
