package polygon

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"

	assetpb "github.com/eolymp/go-sdk/eolymp/asset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type assetMock struct {
	lock   sync.Mutex
	buffer map[string][]byte
}

func (*assetMock) LookupAsset(ctx context.Context, in *assetpb.LookupAssetInput, opts ...grpc.CallOption) (*assetpb.LookupAssetOutput, error) {
	return nil, status.Error(codes.NotFound, "not found")
}

func (*assetMock) UploadAsset(ctx context.Context, in *assetpb.UploadAssetInput, opts ...grpc.CallOption) (*assetpb.UploadAssetOutput, error) {
	hash := fmt.Sprintf("%x", md5.Sum(in.Data))
	return &assetpb.UploadAssetOutput{AssetUrl: "https://eolympusercontent.com/file/" + in.Name + "." + hash}, nil
}

func (m *assetMock) StartMultipartUpload(ctx context.Context, in *assetpb.StartMultipartUploadInput, opts ...grpc.CallOption) (*assetpb.StartMultipartUploadOutput, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.buffer == nil {
		m.buffer = make(map[string][]byte)
	}

	m.buffer[in.GetName()] = []byte{}

	return &assetpb.StartMultipartUploadOutput{UploadId: in.GetName()}, nil
}

func (m *assetMock) UploadPart(ctx context.Context, in *assetpb.UploadPartInput, opts ...grpc.CallOption) (*assetpb.UploadPartOutput, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.buffer[in.GetUploadId()] = append(m.buffer[in.GetUploadId()], in.GetData()...)

	return &assetpb.UploadPartOutput{}, nil
}

func (m *assetMock) CompleteMultipartUpload(ctx context.Context, in *assetpb.CompleteMultipartUploadInput, opts ...grpc.CallOption) (*assetpb.CompleteMultipartUploadOutput, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	hash := fmt.Sprintf("%x", md5.Sum(m.buffer[in.GetUploadId()]))

	return &assetpb.CompleteMultipartUploadOutput{AssetUrl: "https://eolympusercontent.com/file/" + in.GetUploadId() + "." + hash}, nil
}
