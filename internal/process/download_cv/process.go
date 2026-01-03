package download_cv

import "net/http"

type HTTPClient interface {
	ProxyCVDownload(w http.ResponseWriter, token string, lang string) error
}

type Process struct {
	httpClient HTTPClient
}

func NewProcess(client HTTPClient) *Process {
	return &Process{httpClient: client}
}

func (p *Process) Execute(w http.ResponseWriter, token string, lang string) error {
	return p.httpClient.ProxyCVDownload(w, token, lang)
}
