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
	return "", nil
}