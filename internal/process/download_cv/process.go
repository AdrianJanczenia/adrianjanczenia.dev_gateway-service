package download_cv

import (
	"context"
	"net/http"
)

type ContentServiceClient interface {
	ProxyCVDownload(ctx context.Context, w http.ResponseWriter, token string, lang string) error
}

type Process struct {
	contentServiceClient ContentServiceClient
}

func NewProcess(contentServiceClient ContentServiceClient) *Process {
	return &Process{contentServiceClient: contentServiceClient}
}

func (p *Process) Execute(ctx context.Context, w http.ResponseWriter, token string, lang string) error {
	return p.contentServiceClient.ProxyCVDownload(ctx, w, token, lang)
}
