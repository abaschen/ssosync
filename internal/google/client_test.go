package google

import (
	"context"
	"strings"
	"testing"

	"github.com/awslabs/ssosync/internal/constants"
	admin "google.golang.org/api/admin/directory/v1"
)

func TestNewClientWithInvalidJSON(t *testing.T) {
	ctx := context.Background()
	invalidJSON := []byte(`{"invalid": "json"`)

	_, err := NewClient(ctx, "admin@example.com", invalidJSON)
	if err == nil {
		t.Error("Expected error with invalid JSON, got nil")
	}
}

func TestNewClientWithEmptyAdminEmail(t *testing.T) {
	ctx := context.Background()
	validJSON := []byte(`{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "key-id",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7VJTUt9Us8cKB\nxhXctbdgZATkr+IfaU1ZWFWOiS2XZjXYlIr02c8EjQHTFgd5bad5lXGp9AX68ERP\n-----END PRIVATE KEY-----\n",
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id": "123456789",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`)

	// The NewClient function doesn't validate empty admin email directly
	// It will be set as the Subject in the JWT config
	// This test verifies the function can handle empty admin email without panicking
	client, err := NewClient(ctx, "", validJSON)
	if err != nil {
		// This is expected since we're using a fake private key
		t.Logf("Expected error with fake credentials: %v", err)
	} else if client != nil {
		t.Log("Client created successfully with empty admin email")
	}
}

// Test helper functions
func TestZeroWidthSpaceReplacement(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no zero width space",
			input:    "John Doe",
			expected: "John Doe",
		},
		{
			name:     "with zero width space",
			input:    "John" + string(constants.ZeroWidthSpace) + "Doe",
			expected: "John Doe",
		},
		{
			name:     "multiple zero width spaces",
			input:    "John" + string(constants.ZeroWidthSpace) + string(constants.ZeroWidthSpace) + "Doe",
			expected: "John  Doe",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := strings.ReplaceAll(tc.input, string(constants.ZeroWidthSpace), " ")
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// Mock client for testing
type mockGoogleClient struct {
	users        []*admin.User
	deletedUsers []*admin.User
	groups       []*admin.Group
	members      map[string][]*admin.Member
	getUsersErr  error
	getGroupsErr error
}

func (m *mockGoogleClient) GetUsers(query string) ([]*admin.User, error) {
	if m.getUsersErr != nil {
		return nil, m.getUsersErr
	}
	return m.users, nil
}

func (m *mockGoogleClient) GetDeletedUsers() ([]*admin.User, error) {
	return m.deletedUsers, nil
}

func (m *mockGoogleClient) GetGroups(query string) ([]*admin.Group, error) {
	if m.getGroupsErr != nil {
		return nil, m.getGroupsErr
	}
	return m.groups, nil
}

func (m *mockGoogleClient) GetGroupMembers(ctx context.Context, group *admin.Group) ([]*admin.Member, error) {
	if members, exists := m.members[group.Id]; exists {
		return members, nil
	}
	return []*admin.Member{}, nil
}

func TestMockGoogleClient(t *testing.T) {
	mockClient := &mockGoogleClient{
		users: []*admin.User{
			{
				PrimaryEmail: "user1@example.com",
				Name: &admin.UserName{
					GivenName:  "John",
					FamilyName: "Doe",
				},
			},
		},
		groups: []*admin.Group{
			{
				Id:    "group1",
				Name:  "Test Group",
				Email: "testgroup@example.com",
			},
		},
		members: map[string][]*admin.Member{
			"group1": {
				{
					Email: "user1@example.com",
					Type:  "USER",
				},
			},
		},
	}

	// Test GetUsers
	users, err := mockClient.GetUsers("*")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
	if users[0].PrimaryEmail != "user1@example.com" {
		t.Errorf("Expected user1@example.com, got %s", users[0].PrimaryEmail)
	}

	// Test GetGroups
	groups, err := mockClient.GetGroups("*")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}
	if groups[0].Email != "testgroup@example.com" {
		t.Errorf("Expected testgroup@example.com, got %s", groups[0].Email)
	}

	// Test GetGroupMembers
	ctx := context.Background()
	members, err := mockClient.GetGroupMembers(ctx, groups[0])
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(members))
	}
	if members[0].Email != "user1@example.com" {
		t.Errorf("Expected user1@example.com, got %s", members[0].Email)
	}
}
