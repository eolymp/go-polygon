package polygon

import (
	"context"
	assetpb "github.com/eolymp/go-sdk/eolymp/asset"
	"google.golang.org/grpc"
)

type assetMock struct {
}

func (assetMock) UploadFile(ctx context.Context, in *assetpb.UploadFileInput, opts ...grpc.CallOption) (*assetpb.UploadFileOutput, error) {
	return &assetpb.UploadFileOutput{FileUrl: "https://eolympusercontent.com/file/" + in.Name}, nil
}

func (assetMock) UploadImage(ctx context.Context, in *assetpb.UploadImageInput, opts ...grpc.CallOption) (*assetpb.UploadImageOutput, error) {
	return &assetpb.UploadImageOutput{ImageUrl: "https://eolympusercontent.com/" + in.Name}, nil
}
