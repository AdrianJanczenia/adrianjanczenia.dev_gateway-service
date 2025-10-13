package grpc

import (
	"context"

	contentv1 "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
	"google.golang.org/grpc"
)

type Client struct {
	grpc contentv1.ContentServiceClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{grpc: contentv1.NewContentServiceClient(conn)}
}

func (c *Client) GetContent(ctx context.Context, lang string) (string, error) {
	resp, err := c.grpc.Handle(ctx, &contentv1.GetContentRequest{Lang: lang})
	if err != nil {
		return "", err
	}

	return resp.GetJsonContent(), nil
}
