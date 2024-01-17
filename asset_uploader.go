package polygon

import (
	"context"
	assetpb "github.com/eolymp/go-sdk/eolymp/asset"
	"google.golang.org/grpc"
)

type assetUploader interface {
	UploadFile(ctx context.Context, in *assetpb.UploadFileInput, opts ...grpc.CallOption) (*assetpb.UploadFileOutput, error)
	UploadImage(ctx context.Context, in *assetpb.UploadImageInput, opts ...grpc.CallOption) (*assetpb.UploadImageOutput, error)
}
