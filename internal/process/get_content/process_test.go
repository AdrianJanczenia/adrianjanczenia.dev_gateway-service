package get_content

import (
	"context"
	"errors"
	"testing"
)

type mockContentServiceClient struct {
	getContentFunc func(ctx context.Context, lang string) ([]byte, error)
}

func (m *mockContentServiceClient) GetContent(ctx context.Context, lang string) ([]byte, error) {
	return m.getContentFunc(ctx, lang)
}

func TestProcess_GetContent(t *testing.T) {
	tests := []struct {
		name           string
		getContentFunc func(ctx context.Context, lang string) ([]byte, error)
		want           []byte
		wantErr        bool
	}{
		{
			name: "success",
			getContentFunc: func(ctx context.Context, lang string) ([]byte, error) {
				return []byte("data"), nil
			},
			want:    []byte("data"),
			wantErr: false,
		},
		{
			name: "error from service",
			getContentFunc: func(ctx context.Context, lang string) ([]byte, error) {
				return nil, errors.New("error")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockContentServiceClient{getContentFunc: tt.getContentFunc}
			p := NewProcess(m)
			got, err := p.Process(context.Background(), "en")
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if string(got) != string(tt.want) {
				t.Errorf("Process() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
