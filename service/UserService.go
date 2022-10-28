package service

import "sca_server/container"

type UserService interface {
	LoginMethod()
}

type userService struct {
	container container.Container
}

func (u userService) LoginMethod() {
	panic("implement me")
}

// NewUserService is constructor.
func NewUserService(container container.Container) UserService {
	return &userService{container: container}
}
