package grpc

import (
	"context"

	pb "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client pb.ContentServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{client: pb.NewContentServiceClient(conn)}
}

func (c *Client) GetContent(ctx context.Context, lang string) ([]byte, error) {
	resp, err := c.client.Handle(ctx, &pb.GetContentRequest{Lang: lang})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, errors.ErrContentNotFound
		}
		return nil, errors.ErrServiceUnavailable
	}
	return []byte(resp.JsonContent), nil
}
