package usecase_test

import (
	"context"
	"errors"
	v1 "github.com/k4l1ma/EchoSphere/api/v1"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/core"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase"
	"github.com/k4l1ma/EchoSphere/internal/EchoSphereServer/usecase/internal/mocks"
	"go.uber.org/mock/gomock"
)

func (u *useCaseSuite) TestRelayHandler_Success() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := usecase.RelayCMD{From: "owner1", Content: "test message"}
	randomOwnerID := "random1"

	ownerRelayer := mocks.NewMockMessager(gomock.NewController(u.T()))

	randomRelayer := mocks.NewMockMessager(gomock.NewController(u.T()))

	u.router.EXPECT().AcquireRandomRelayer(ctx, cmd.From).Return(randomOwnerID, randomRelayer, nil)

	u.router.EXPECT().ReleaseRelayer(ctx, randomOwnerID, randomRelayer)

	randomRelayer.EXPECT().SendMsg(
		ctx,
		&v1.EchoSphereTransmissionServiceTransmitResponse{
			OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Message{
				Message: &v1.Message{From: cmd.From, Content: cmd.Content},
			},
		},
	).Return(nil)

	u.router.EXPECT().AcquireRelayer(ctx, cmd.From).Return(ownerRelayer, nil)

	u.router.EXPECT().ReleaseRelayer(ctx, cmd.From, ownerRelayer)

	ownerRelayer.EXPECT().SendMsg(
		ctx,
		&v1.EchoSphereTransmissionServiceTransmitResponse{
			OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Message{
				Message: &v1.Message{From: cmd.From, Content: cmd.Content},
			},
		},
	).Return(nil)

	err := u.SUT.RelayHandler(ctx, cmd)
	u.Require().NoError(err)
}

func (u *useCaseSuite) TestRelayHandler_RelayerNotFound() {
	ctx := context.Background()
	cmd := usecase.RelayCMD{From: "owner1", Content: "test message"}

	u.router.EXPECT().AcquireRandomRelayer(ctx, cmd.From).Return("", nil, errors.New("relayer not found"))

	err := u.SUT.RelayHandler(ctx, cmd)
	u.Require().Error(err)
}

func (u *useCaseSuite) TestRelayHandler_FailedToGetRandomRelayer() {
	ctx := context.Background()
	cmd := usecase.RelayCMD{From: "owner1", Content: "test message"}
	ownerRelayer := mocks.NewMockMessager(gomock.NewController(u.T()))
	NoOpRelayer := mocks.NewMockMessager(gomock.NewController(u.T()))

	NoOpRelayer.EXPECT().SendMsg(gomock.Any(), gomock.Any()).Return(nil)

	u.router.
		EXPECT().
		AcquireRandomRelayer(ctx, cmd.From).
		Return("ffffffff-ffff-ffff-ffff-ffffffffffff", NoOpRelayer, core.ErrFailedToGetRelayer)

	u.router.EXPECT().AcquireRelayer(ctx, cmd.From).Return(ownerRelayer, nil)

	u.router.EXPECT().ReleaseRelayer(ctx, gomock.Any(), gomock.Any()).Times(2)

	ownerRelayer.
		EXPECT().
		SendMsg(
			ctx,
			&v1.EchoSphereTransmissionServiceTransmitResponse{
				OutgoingData: &v1.EchoSphereTransmissionServiceTransmitResponse_Message{
					Message: &v1.Message{From: cmd.From, Content: cmd.Content},
				},
			},
		).
		Return(nil)

	err := u.SUT.RelayHandler(ctx, cmd)
	u.Require().NoError(err)
}
