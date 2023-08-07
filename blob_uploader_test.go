package polygon

import (
	"context"
	keeperpb "github.com/eolymp/go-sdk/eolymp/keeper"
	"google.golang.org/grpc"
)

type blobMock struct {
}

func (blobMock) StartMultipartUpload(ctx context.Context, in *keeperpb.StartMultipartUploadInput, opts ...grpc.CallOption) (*keeperpb.StartMultipartUploadOutput, error) {
	return &keeperpb.StartMultipartUploadOutput{ObjectId: "mocked-object-id"}, nil
}

func (blobMock) UploadPart(ctx context.Context, in *keeperpb.UploadPartInput, opts ...grpc.CallOption) (*keeperpb.UploadPartOutput, error) {
	return &keeperpb.UploadPartOutput{}, nil
}

func (blobMock) CompleteMultipartUpload(ctx context.Context, in *keeperpb.CompleteMultipartUploadInput, opts ...grpc.CallOption) (*keeperpb.CompleteMultipartUploadOutput, error) {
	return &keeperpb.CompleteMultipartUploadOutput{}, nil
}
