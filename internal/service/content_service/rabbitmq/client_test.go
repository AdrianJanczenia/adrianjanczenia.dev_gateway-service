package rabbitmq

import (
	"encoding/json"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

func TestRABBITMQClient_ResponseParsing(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		wantErr error
	}{
		{
			name:    "valid response",
			body:    []byte(`{"token":"abc"}`),
			wantErr: nil,
		},
		{
			name:    "error slug in response",
			body:    []byte(`{"error":"error_cv_auth"}`),
			wantErr: errors.ErrInvalidPassword,
		},
		{
			name:    "unknown error slug",
			body:    []byte(`{"error":"unknown_something"}`),
			wantErr: errors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp struct {
				Error string `json:"error"`
			}
			json.Unmarshal(tt.body, &resp)

			var err error
			if resp.Error != "" {
				err = errors.FromSlug(resp.Error)
			}

			if err != tt.wantErr {
				t.Errorf("Parsing error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
