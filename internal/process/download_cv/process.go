package download_cv

import "net/http"

type ContentServiceClient interface {
	ProxyCVDownload(w http.ResponseWriter, token string, lang string) error
}

type Process struct {
	contentServiceClient ContentServiceClient
}

func NewProcess(contentServiceClient ContentServiceClient) *Process {
	return &Process{contentServiceClient: contentServiceClient}
}

func (p *Process) Execute(w http.ResponseWriter, token string, lang string) error {
	return p.contentServiceClient.ProxyCVDownload(w, token, lang)
}
