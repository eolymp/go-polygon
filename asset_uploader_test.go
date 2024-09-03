package polygon

import (
	"context"
	assetpb "github.com/eolymp/go-sdk/eolymp/asset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type assetMock struct {
}

func (assetMock) ResolveAlias(ctx context.Context, in *assetpb.ResolveAliasInput, opts ...grpc.CallOption) (*assetpb.ResolveAliasOutput, error) {
	return nil, status.Error(codes.NotFound, "not found")
}

func (assetMock) UploadAsset(ctx context.Context, in *assetpb.UploadAssetInput, opts ...grpc.CallOption) (*assetpb.UploadAssetOutput, error) {
	return &assetpb.UploadAssetOutput{AssetUrl: "https://eolympusercontent.com/file/" + in.Name}, nil
}

func (assetMock) StartMultipartUpload(ctx context.Context, in *assetpb.StartMultipartUploadInput, opts ...grpc.CallOption) (*assetpb.StartMultipartUploadOutput, error) {
	return &assetpb.StartMultipartUploadOutput{UploadId: in.GetName()}, nil
}

func (assetMock) UploadPart(ctx context.Context, in *assetpb.UploadPartInput, opts ...grpc.CallOption) (*assetpb.UploadPartOutput, error) {
	return &assetpb.UploadPartOutput{}, nil
}

func (assetMock) CompleteMultipartUpload(ctx context.Context, in *assetpb.CompleteMultipartUploadInput, opts ...grpc.CallOption) (*assetpb.CompleteMultipartUploadOutput, error) {
	return &assetpb.CompleteMultipartUploadOutput{AssetUrl: in.GetUploadId()}, nil
}
