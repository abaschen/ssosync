package config

import (
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				GoogleAdmin:     "admin@example.com",
				SCIMEndpoint:    "https://scim.example.com",
				SCIMAccessToken: "token123",
				Region:          "us-east-1",
				IdentityStoreID: "d-123456789",
				SyncMethod:      "groups",
			},
			expectError: false,
		},
		{
			name: "missing google admin",
			config: &Config{
				SCIMEndpoint:    "https://scim.example.com",
				SCIMAccessToken: "token123",
				Region:          "us-east-1",
				IdentityStoreID: "d-123456789",
				SyncMethod:      "groups",
			},
			expectError: true,
			errorMsg:    "google admin email is required",
		},
		{
			name: "missing SCIM endpoint",
			config: &Config{
				GoogleAdmin:     "admin@example.com",
				SCIMAccessToken: "token123",
				Region:          "us-east-1",
				IdentityStoreID: "d-123456789",
				SyncMethod:      "groups",
			},
			expectError: true,
			errorMsg:    "SCIM endpoint is required",
		},
		{
			name: "invalid sync method",
			config: &Config{
				GoogleAdmin:     "admin@example.com",
				SCIMEndpoint:    "https://scim.example.com",
				SCIMAccessToken: "token123",
				Region:          "us-east-1",
				IdentityStoreID: "d-123456789",
				SyncMethod:      "invalid",
			},
			expectError: true,
			errorMsg:    "sync method must be either 'groups' or 'users_groups'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
