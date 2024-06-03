package usecase_test

import (
	"context"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"go.uber.org/mock/gomock"
)

type mockStreamSender struct{}

func (mockStreamSender) SendMsg(context.Context, any) error { return nil }

func (u *useCaseSuite) TestRegisterHandler() {
	ctx := context.Background()
	cmd := usecase.RegisterCMD{
		OwnerID:      "client-1",
		StreamSender: mockStreamSender{},
	}

	u.router.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	err := u.SUT.RegisterHandler(ctx, cmd)
	u.Require().NoError(err)
}
