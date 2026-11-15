package handler

import (
	"context"

	modelv1 "github.com/thetnaingtn/ai-system/gen/go/platform/model/v1"
)

// ModelHandler provides placeholders for the model-management RPCs.
type ModelHandler struct {
	modelv1.UnimplementedModelServiceServer
}

var _ modelv1.ModelServiceServer = (*ModelHandler)(nil)

func (*ModelHandler) RegisterModel(context.Context, *modelv1.RegisterModelRequest) (*modelv1.RegisterModelResponse, error) {
	panic("not implemented")
}

func (*ModelHandler) GetModel(context.Context, *modelv1.GetModelRequest) (*modelv1.GetModelResponse, error) {
	panic("not implemented")
}

func (*ModelHandler) ListModels(context.Context, *modelv1.ListModelsRequest) (*modelv1.ListModelsResponse, error) {
	panic("not implemented")
}

func (*ModelHandler) UpdateModel(context.Context, *modelv1.UpdateModelRequest) (*modelv1.UpdateModelResponse, error) {
	panic("not implemented")
}

func (*ModelHandler) GetModelChange(context.Context, *modelv1.GetModelChangeRequest) (*modelv1.GetModelChangeResponse, error) {
	panic("not implemented")
}

func (*ModelHandler) ListModelChanges(context.Context, *modelv1.ListModelChangesRequest) (*modelv1.ListModelChangesResponse, error) {
	panic("not implemented")
}

func (*ModelHandler) ApproveModelChange(context.Context, *modelv1.ApproveModelChangeRequest) (*modelv1.ApproveModelChangeResponse, error) {
	panic("not implemented")
}

func (*ModelHandler) RejectModelChange(context.Context, *modelv1.RejectModelChangeRequest) (*modelv1.RejectModelChangeResponse, error) {
	panic("not implemented")
}

func (*ModelHandler) RotateModelCredential(context.Context, *modelv1.RotateModelCredentialRequest) (*modelv1.RotateModelCredentialResponse, error) {
	panic("not implemented")
}
