// Copyright (c) 2020, Amazon.com, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	u := NewUser("Lee", "Packham", "test@email.com", true)
	assert.Equal(t, "Lee", u.Name.GivenName)
	assert.Equal(t, "Packham", u.Name.FamilyName)
	assert.Equal(t, "Lee Packham", u.DisplayName)
	assert.Len(t, u.Emails, 1)

	assert.Equal(t, "test@email.com", u.Emails[0].Value)
	assert.True(t, u.Emails[0].Primary)

	assert.True(t, u.Active)

	assert.Len(t, u.Schemas, 1)
	assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User", u.Schemas[0])
}

func TestUpdateUser(t *testing.T) {
	u := UpdateUser("111", "Lee", "Packham", "test@email.com", false)
	assert.Equal(t, "Lee", u.Name.GivenName)
	assert.Equal(t, "Packham", u.Name.FamilyName)
	assert.Equal(t, "Lee Packham", u.DisplayName)
	assert.Len(t, u.Emails, 1)

	assert.Equal(t, "test@email.com", u.Emails[0].Value)
	assert.True(t, u.Emails[0].Primary)

	assert.False(t, u.Active)

	assert.Len(t, u.Schemas, 1)
	assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User", u.Schemas[0])
}
