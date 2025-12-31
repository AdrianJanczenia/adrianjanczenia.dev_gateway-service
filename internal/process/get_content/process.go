package get_content

import "context"

type ContentService interface {
	GetContent(ctx context.Context, lang string) ([]byte, error)
}

type Process struct {
	contentService ContentService
}

func NewProcess(contentService ContentService) *Process {
	return &Process{contentService: contentService}
}

func (p *Process) Process(ctx context.Context, lang string) ([]byte, error) {
	return p.contentService.GetContent(ctx, lang)
}
