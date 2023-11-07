package polygon

import (
	"context"
	assetpb "github.com/eolymp/go-sdk/eolymp/asset"
	"google.golang.org/grpc"
)

type assetMock struct {
}

func (assetMock) UploadFile(ctx context.Context, in *assetpb.UploadFileInput, opts ...grpc.CallOption) (*assetpb.UploadFileOutput, error) {
	return nil, nil
}
