package grpc

import (
	"context"
	"errors"
	"testing"

	pb "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockContentServiceClient struct {
	pb.ContentServiceClient
	handleFunc func(ctx context.Context, in *pb.GetContentRequest, opts ...grpc.CallOption) (*pb.GetContentResponse, error)
}

func (m *mockContentServiceClient) Handle(ctx context.Context, in *pb.GetContentRequest, opts ...grpc.CallOption) (*pb.GetContentResponse, error) {
	return m.handleFunc(ctx, in, opts...)
}

func TestGRPCClient_GetContent(t *testing.T) {
	tests := []struct {
		name       string
		handleFunc func(context.Context, *pb.GetContentRequest, ...grpc.CallOption) (*pb.GetContentResponse, error)
		want       []byte
		wantErr    error
	}{
		{
			name: "successful response",
			handleFunc: func(ctx context.Context, in *pb.GetContentRequest, opts ...grpc.CallOption) (*pb.GetContentResponse, error) {
				return &pb.GetContentResponse{JsonContent: `{"key":"value"}`}, nil
			},
			want:    []byte(`{"key":"value"}`),
			wantErr: nil,
		},
		{
			name: "not found error",
			handleFunc: func(ctx context.Context, in *pb.GetContentRequest, opts ...grpc.CallOption) (*pb.GetContentResponse, error) {
				return nil, status.Error(codes.NotFound, "not found")
			},
			want:    nil,
			wantErr: appErrors.ErrContentNotFound,
		},
		{
			name: "internal server error",
			handleFunc: func(ctx context.Context, in *pb.GetContentRequest, opts ...grpc.CallOption) (*pb.GetContentResponse, error) {
				return nil, errors.New("grpc error")
			},
			want:    nil,
			wantErr: appErrors.ErrServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockContentServiceClient{handleFunc: tt.handleFunc}
			c := &Client{client: m}
			got, err := c.GetContent(context.Background(), "pl")
			if err != tt.wantErr {
				t.Errorf("GetContent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if string(got) != string(tt.want) {
				t.Errorf("GetContent() got = %s, want %s", string(got), string(tt.want))
			}
		})
	}
}
