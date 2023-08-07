package polygon

import (
	"context"
	keeperpb "github.com/eolymp/go-sdk/eolymp/keeper"
	"google.golang.org/grpc"
)

type blobUploader interface {
	StartMultipartUpload(ctx context.Context, in *keeperpb.StartMultipartUploadInput, opts ...grpc.CallOption) (*keeperpb.StartMultipartUploadOutput, error)
	UploadPart(ctx context.Context, in *keeperpb.UploadPartInput, opts ...grpc.CallOption) (*keeperpb.UploadPartOutput, error)
	CompleteMultipartUpload(ctx context.Context, in *keeperpb.CompleteMultipartUploadInput, opts ...grpc.CallOption) (*keeperpb.CompleteMultipartUploadOutput, error)
}
