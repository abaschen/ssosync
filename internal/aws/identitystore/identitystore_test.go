package identitystore

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/awslabs/ssosync/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMemberIdMemberUserId(t *testing.T) {
	tests := []struct {
		name        string
		memberId    types.MemberId
		expected    *string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid user id",
			memberId: &types.MemberIdMemberUserId{
				Value: "user-123",
			},
			expected:    stringPtr("user-123"),
			expectError: false,
		},
		{
			name:        "unknown union member",
			memberId:    &types.UnknownUnionMember{Tag: "unknown", Value: []byte("test")},
			expected:    nil,
			expectError: true,
			errorMsg:    "expected a user id, got unknown type id",
		},
		{
			name:        "nil member id",
			memberId:    nil,
			expected:    nil,
			expectError: true,
			errorMsg:    "expected a user id, got unknown type id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetMemberIdMemberUserId(tt.memberId)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestIsMemberInGroups(t *testing.T) {
	tests := []struct {
		name            string
		identityStoreId string
		groupIds        []string
		memberId        string
		mockResponse    *identitystore.IsMemberInGroupsOutput
		mockError       error
		expected        *bool
		expectError     bool
	}{
		{
			name:            "user is member of group",
			identityStoreId: "d-123456789",
			groupIds:        []string{"group-123"},
			memberId:        "user-123",
			mockResponse: &identitystore.IsMemberInGroupsOutput{
				Results: []types.GroupMembershipExistenceResult{
					{
						GroupId:          stringPtr("group-123"),
						MemberId:         &types.MemberIdMemberUserId{Value: "user-123"},
						MembershipExists: true,
					},
				},
			},
			mockError:   nil,
			expected:    boolPtr(true),
			expectError: false,
		},
		{
			name:            "user is not member of group",
			identityStoreId: "d-123456789",
			groupIds:        []string{"group-123"},
			memberId:        "user-123",
			mockResponse: &identitystore.IsMemberInGroupsOutput{
				Results: []types.GroupMembershipExistenceResult{
					{
						GroupId:          stringPtr("group-123"),
						MemberId:         &types.MemberIdMemberUserId{Value: "user-123"},
						MembershipExists: false,
					},
				},
			},
			mockError:   nil,
			expected:    boolPtr(false),
			expectError: false,
		},
		{
			name:            "API error",
			identityStoreId: "d-123456789",
			groupIds:        []string{"group-123"},
			memberId:        "user-123",
			mockResponse:    nil,
			mockError:       errors.New("API error"),
			expected:        nil,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := mocks.NewMockIdentityStoreAPI(t)

			if tt.mockError != nil {
				mockAPI.EXPECT().IsMemberInGroups(mock.Anything, mock.Anything).Return(nil, tt.mockError)
			} else {
				mockAPI.EXPECT().IsMemberInGroups(mock.Anything, mock.Anything).Return(tt.mockResponse, nil)
			}

			ctx := context.Background()
			result, err := IsMemberInGroups(ctx, mockAPI, &tt.identityStoreId, tt.groupIds, &tt.memberId)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestGetGroupMembershipId(t *testing.T) {
	tests := []struct {
		name            string
		identityStoreId string
		groupId         string
		memberId        string
		mockResponse    *identitystore.GetGroupMembershipIdOutput
		mockError       error
		expectError     bool
	}{
		{
			name:            "successful get membership id",
			identityStoreId: "d-123456789",
			groupId:         "group-123",
			memberId:        "user-123",
			mockResponse: &identitystore.GetGroupMembershipIdOutput{
				MembershipId:    stringPtr("membership-123"),
				IdentityStoreId: stringPtr("d-123456789"),
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:            "API error",
			identityStoreId: "d-123456789",
			groupId:         "group-123",
			memberId:        "user-123",
			mockResponse:    nil,
			mockError:       errors.New("API error"),
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := mocks.NewMockIdentityStoreAPI(t)

			if tt.mockError != nil {
				mockAPI.EXPECT().GetGroupMembershipId(mock.Anything, mock.Anything).Return(nil, tt.mockError)
			} else {
				mockAPI.EXPECT().GetGroupMembershipId(mock.Anything, mock.Anything).Return(tt.mockResponse, nil)
			}

			ctx := context.Background()
			result, err := GetGroupMembershipId(ctx, mockAPI, &tt.identityStoreId, &tt.groupId, &tt.memberId)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponse, result)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
