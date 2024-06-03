package usecase_test

import (
	"context"
	"errors"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase/internal/mocks"
	"go.uber.org/mock/gomock"
)

func (u *useCaseSuite) TestAckHandler_Success() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := usecase.AckCMD{
		From:    "client-1",
		To:      "client-2",
		Content: "ack message",
	}

	ownerRelayer := mocks.NewMockMessager(gomock.NewController(u.T()))

	u.router.EXPECT().AcquireRelayer(ctx, cmd.To).Return(ownerRelayer, nil)
	u.router.EXPECT().ReleaseRelayer(ctx, cmd.To, ownerRelayer)
	ownerRelayer.EXPECT().SendMsg(
		ctx,
		&v1.EchoSphereTransmissionServiceTransmitResponse{
			OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Ack{
				Ack: &v1.Ack{From: cmd.From, To: cmd.To, Content: cmd.Content},
			},
		},
	).Return(nil)

	err := u.SUT.AckHandler(ctx, cmd)
	u.NoError(err)
}

func (u *useCaseSuite) TestAckHandler_RelayerNotFound() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := usecase.AckCMD{
		From:    "client-1",
		To:      "client-2",
		Content: "ack message",
	}

	u.router.EXPECT().AcquireRelayer(ctx, cmd.To).Return(nil, errors.New("relayer not found"))

	err := u.SUT.AckHandler(ctx, cmd)
	u.Error(err)
}
