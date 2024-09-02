package polygon

import (
	"context"
	assetpb "github.com/eolymp/go-sdk/eolymp/asset"
	"google.golang.org/grpc"
)

type blobMock struct {
}

func (blobMock) StartMultipartUpload(ctx context.Context, in *assetpb.StartMultipartUploadInput, opts ...grpc.CallOption) (*assetpb.StartMultipartUploadOutput, error) {
	return &assetpb.StartMultipartUploadOutput{UploadId: in.GetName()}, nil
}

func (blobMock) UploadPart(ctx context.Context, in *assetpb.UploadPartInput, opts ...grpc.CallOption) (*assetpb.UploadPartOutput, error) {
	return &assetpb.UploadPartOutput{}, nil
}

func (blobMock) CompleteMultipartUpload(ctx context.Context, in *assetpb.CompleteMultipartUploadInput, opts ...grpc.CallOption) (*assetpb.CompleteMultipartUploadOutput, error) {
	return &assetpb.CompleteMultipartUploadOutput{AssetUrl: in.GetUploadId()}, nil
}
