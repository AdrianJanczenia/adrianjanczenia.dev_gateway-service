package get_content

import "context"

type ContentServiceClient interface {
	GetContent(ctx context.Context, lang string) ([]byte, error)
}

type Process struct {
	contentServiceClient ContentServiceClient
}

func NewProcess(contentServiceClient ContentServiceClient) *Process {
	return &Process{contentServiceClient: contentServiceClient}
}

func (p *Process) Process(ctx context.Context, lang string) ([]byte, error) {
	return p.contentServiceClient.GetContent(ctx, lang)
}
