package user

import "context"

type AuthDataContainer interface {
}

type userAuthorizer struct {
	dataContainer AuthDataContainer
}

func NewAuthorizer(dc AuthDataContainer) *userAuthorizer {
	return &userAuthorizer{dataContainer: dc}
}

func (u *userAuthorizer) Register(ctx context.Context, login, password string) (string, error) {
	//TODO
	return "Bearer 1234abcd", nil
}

func (u *userAuthorizer) Login(ctx context.Context, login, password string) (string, error) {
	//TODO
	return "Bearer 1234abcd", nil
}

func (u *userAuthorizer) AuthByToken(ctx context.Context, accessToken string) (userID string, err error) {
	//TODO
	return "aaaa-bbbb-cccc-dddd", nil
}
