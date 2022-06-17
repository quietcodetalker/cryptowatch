package user_test

import (
	"context"
	"cryptowatch/internal/app/user"
	"cryptowatch/internal/app/user/mock"
	"cryptowatch/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	password := "password1"
	passwordHash, err := util.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, passwordHash)

	u := user.User{
		ID:           1,
		Username:     "user1",
		PasswordHash: passwordHash,
		FirstName:    "fname1",
		LastName:     "lname1",
		CreateTime:   time.Now(),
	}

	tests := []struct {
		name       string
		buildStubs func(repo *mock.MockRepository)
		req        user.SvcCreateReq
		res        *user.User
		err        error
	}{
		{
			name: "OK",
			buildStubs: func(repo *mock.MockRepository) {
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&u, nil)
			},
			req: user.SvcCreateReq{
				Username:  u.Username,
				Password:  password,
				FirstName: u.FirstName,
				LastName:  u.LastName,
			},
			res: &u,
			err: nil,
		},
		{
			name: "Internal error",
			buildStubs: func(repo *mock.MockRepository) {
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, user.ErrInternalError)
			},
			req: user.SvcCreateReq{
				Username:  u.Username,
				Password:  password,
				FirstName: u.FirstName,
				LastName:  u.LastName,
			},
			res: nil,
			err: user.ErrInternalError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mock.NewMockRepository(ctrl)
			tt.buildStubs(repo)

			svc := user.NewService(repo, nil, nil)

			res, err := svc.Create(context.Background(), tt.req)
			assert.ErrorIs(t, err, tt.err)
			if tt.res == nil {
				assert.Nil(t, res)
			} else {
				require.NotNil(t, res)

				assert.Equal(t, tt.res.ID, res.ID)
				assert.Equal(t, tt.res.Username, res.Username)
				assert.Equal(t, tt.res.FirstName, res.FirstName)
				assert.Equal(t, tt.res.LastName, res.LastName)

				assert.NotEmpty(t, res.PasswordHash)

				assert.WithinDuration(t, tt.res.CreateTime, res.CreateTime, time.Second)
			}
		})
	}
}
