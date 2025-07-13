package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthenticator struct {
	mock.Mock
}

func (m *MockAuthenticator) AuthTokenForHost(host string) (string, error) {
	args := m.Called(host)
	return args.String(0), args.Error(1)
}

func TestNewClient(t *testing.T) {
	mockAuth := new(MockAuthenticator)
	mockAuth.On("AuthTokenForHost", mock.Anything).Return("dummy-token", nil)

	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{
			name:    "Default host",
			host:    "github.com",
			wantErr: false,
		},
		{
			name:    "Custom host",
			host:    "custom.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.host, mockAuth, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if client == nil {
				t.Errorf("Expected non-nil client")
				return
			}
			assert.NotNil(t, client.client, "The github client should not be nil")
		})
	}
}
