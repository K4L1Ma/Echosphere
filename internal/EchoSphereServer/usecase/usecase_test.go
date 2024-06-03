package usecase_test

import (
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase/internal/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
)

//go:generate go run go.uber.org/mock/mockgen@latest  -destination=./internal/mocks/multiplexer.mock.go -package=mocks -source=../core/router.go RelayRouter

type useCaseSuite struct {
	suite.Suite

	router *mocks.MockRelayRouter
	SUT    *usecase.UC
}

func (u *useCaseSuite) SetupSuite() {
	ctrl := gomock.NewController(u.T())

	u.router = mocks.NewMockRelayRouter(ctrl)

	u.SUT = usecase.New(u.router)
}

func TestUseCases(t *testing.T) {
	suite.Run(t, new(useCaseSuite))
}
