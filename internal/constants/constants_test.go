package constants

import (
	"net/http"
	"testing"
)

func TestConstants(t *testing.T) {
	// Test SCIM Schema constants
	if SCIMSchemaUser != "urn:ietf:params:scim:schemas:core:2.0:User" {
		t.Errorf("Expected SCIMSchemaUser to be 'urn:ietf:params:scim:schemas:core:2.0:User', got %s", SCIMSchemaUser)
	}

	if SCIMSchemaGroup != "urn:ietf:params:scim:schemas:core:2.0:Group" {
		t.Errorf("Expected SCIMSchemaGroup to be 'urn:ietf:params:scim:schemas:core:2.0:Group', got %s", SCIMSchemaGroup)
	}

	// Test HTTP Status Code constants
	if StatusConflict != http.StatusConflict {
		t.Errorf("Expected StatusConflict to be %d, got %d", http.StatusConflict, StatusConflict)
	}

	// Test Content Type constants
	if ContentTypeSCIM != "application/scim+json" {
		t.Errorf("Expected ContentTypeSCIM to be 'application/scim+json', got %s", ContentTypeSCIM)
	}

	// Test Unicode constants
	if ZeroWidthSpace != '\u200B' {
		t.Errorf("Expected ZeroWidthSpace to be '\\u200B', got %c", ZeroWidthSpace)
	}
}
