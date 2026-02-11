package polygon

import (
	"context"

	assetpb "github.com/eolymp/go-sdk/eolymp/asset"
	"google.golang.org/grpc"
)

type assetUploader interface {
	LookupAsset(ctx context.Context, in *assetpb.LookupAssetInput, opts ...grpc.CallOption) (*assetpb.LookupAssetOutput, error)
	UploadAsset(ctx context.Context, in *assetpb.UploadAssetInput, opts ...grpc.CallOption) (*assetpb.UploadAssetOutput, error)
	StartMultipartUpload(ctx context.Context, in *assetpb.StartMultipartUploadInput, opts ...grpc.CallOption) (*assetpb.StartMultipartUploadOutput, error)
	UploadPart(ctx context.Context, in *assetpb.UploadPartInput, opts ...grpc.CallOption) (*assetpb.UploadPartOutput, error)
	CompleteMultipartUpload(ctx context.Context, in *assetpb.CompleteMultipartUploadInput, opts ...grpc.CallOption) (*assetpb.CompleteMultipartUploadOutput, error)
}
