package polygon

import (
	"context"
	assetservice "github.com/eolymp/go-sdk/eolymp/asset"
	"google.golang.org/grpc"
)

type assetUploader interface {
	UploadFile(ctx context.Context, in *assetservice.UploadFileInput, opts ...grpc.CallOption) (*assetservice.UploadFileOutput, error)
}
