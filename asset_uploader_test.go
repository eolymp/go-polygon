package polygon

import (
	"context"
	typewriterpb "github.com/eolymp/go-sdk/eolymp/typewriter"
	"google.golang.org/grpc"
)

type assetMock struct {
}

func (assetMock) UploadAsset(ctx context.Context, in *typewriterpb.UploadAssetInput, opts ...grpc.CallOption) (*typewriterpb.UploadAssetOutput, error) {
	return nil, nil
}
