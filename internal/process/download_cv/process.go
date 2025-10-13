package download_cv

import "net/http"

type HTTPClient interface {
	ProxyCVDownload(w http.ResponseWriter, r *http.Request) error
}

type Process struct {
	httpClient HTTPClient
}

func NewProcess(client HTTPClient) *Process {
	return &Process{httpClient: client}
}

func (p *Process) Execute(w http.ResponseWriter, r *http.Request) error {
	return p.httpClient.ProxyCVDownload(w, r)
}
