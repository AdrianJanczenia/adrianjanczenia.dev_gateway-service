package get_content

import "context"

type GRPCClient interface {
	GetContent(ctx context.Context, lang string) (string, error)
}

type Process struct {
	grpcClient GRPCClient
}

func NewProcess(client GRPCClient) *Process {
	return &Process{grpcClient: client}
}

func (p *Process) Execute(lang string) (string, error) {
	return p.grpcClient.GetContent(context.Background(), lang)
}
