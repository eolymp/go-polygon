package polygon

import (
	"context"
	assetservice "github.com/eolymp/go-sdk/eolymp/asset"
	"google.golang.org/grpc"
)

type assetMock struct {
}

func (assetMock) UploadAsset(ctx context.Context, in *assetservice.UploadFileInput, opts ...grpc.CallOption) (*assetservice.UploadFileOutput, error) {
	return nil, nil
}
