package service

import "sca_server/container"

type IndexService interface {
	IndexMethod()
}

type indexService struct {
	container container.Container
}

func (u indexService) IndexMethod() {
	panic("implement me")
}

// NewUserService is constructor.
func NewIndexService(container container.Container) UserService {
	return &userService{container: container}
}
