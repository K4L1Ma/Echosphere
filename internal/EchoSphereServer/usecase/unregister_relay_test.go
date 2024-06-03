package usecase_test

import (
	"context"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"go.uber.org/mock/gomock"
)

func (u *useCaseSuite) TestUnregisterHandler() {
	ctx := context.Background()
	cmd := usecase.UnregisterCMD{
		OwnerID: "client-1",
	}

	u.router.EXPECT().AcquireRelayer(gomock.Any(), gomock.Any()).Times(1)

	err := u.SUT.UnregisterHandler(ctx, cmd)
	u.Require().NoError(err)
}
