package grpc

import (
	"context"

	pb "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
	"google.golang.org/grpc"
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
		return nil, err
	}
	return []byte(resp.JsonContent), nil
}
