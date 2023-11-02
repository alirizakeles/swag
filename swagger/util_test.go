// Copyright 2017 Matt Ho
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package swagger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Mock struct {
	name string
	pkg  string
	str  string
}

func (m Mock) PkgPath() string {
	return m.pkg
}

func (m Mock) Name() string {
	return m.name
}

func (m Mock) String() string {
	return m.str
}

func TestMakeSchema(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		str                     string
		pkg                     string
		stripPrefixes           []string
		expectedName            string
		expectedNameWithPackage string
	}{
		{
			"Pet",
			"",
			"package-name",
			[]string{},
			"Pet",
			"package_name_Pet",
		},
		{
			"",
			"string",
			"",
			[]string{},
			"string",
			"string",
		},
		{
			"",
			"int32",
			"package-name",
			[]string{},
			"int32",
			"int32",
		},
		{
			"Pet[string]",
			"",
			"",
			[]string{},
			"Pet[string]",
			"Pet[string]",
		},
		{
			"Pet[string]",
			"",
			"package-name",
			[]string{},
			"Pet[string]",
			"package_name_Pet[string]",
		},
		{
			"Pet[uint64]",
			"",
			"gitlab.com/some-ORG/repo-name/types",
			[]string{"gitlab.com/some-ORG/"},
			"Pet[uint64]",
			"repo_name/types_Pet[uint64]",
		},
		{
			"Pet",
			"",
			"",
			[]string{},
			"Pet",
			"Pet",
		},
		{
			"Pet",
			"",
			"gitlab.com/some-ORG/repo-name/types",
			[]string{},
			"Pet",
			"gitlab_com/some_ORG/repo_name/types_Pet",
		},
		{
			"Pet",
			"",
			"gitlab.com/some-ORG/repo-name/types",
			[]string{"gitlab.com/some-ORG/", "gitlab.com/some-other-ORG/"},
			"Pet",
			"repo_name/types_Pet",
		},
		{
			"Pet",
			"",
			"gitlab.com/some-ORG/repo-name/types",
			[]string{"gitlab.com/some-other-ORG/"},
			"Pet",
			"gitlab_com/some_ORG/repo_name/types_Pet",
		},
		{
			"Pet[A]",
			"",
			"gitlab.com/some-ORG/repo-name/types",
			[]string{"gitlab.com/some-ORG/", "gitlab.com/some-other-ORG/"},
			"Pet[A]",
			"repo_name/types_Pet[repo_name/types_A]",
		},
		{
			"Pet[A, B]",
			"",
			"gitlab.com/some-ORG/repo-name/types",
			[]string{"gitlab.com/some-ORG/", "gitlab.com/some-other-ORG/"},
			"Pet[A, B]",
			"repo_name/types_Pet[repo_name/types_A, repo_name/types_B]",
		},
		{
			"Pet[gitlab.com/some-ORG/other-repo-name/types.A, gitlab.com/some-other-ORG/repo-name/types.B]",
			"",
			"gitlab.com/some-ORG/repo-name/types",
			[]string{"gitlab.com/some-ORG/", "gitlab.com/some-other-ORG/"},
			"Pet[other_repo_name/types_A, repo_name/types_B]", // both not local types, one from other repo, one from other ORG
			"repo_name/types_Pet[other_repo_name/types_A, repo_name/types_B]",
		},
		{
			"Pet[gitlab.com/some-ORG/other-repo-name/types.A, gitlab.com/some-ORG/repo-name/types.B]",
			"",
			"gitlab.com/some-ORG/repo-name/types",
			[]string{"gitlab.com/some-ORG/", "gitlab.com/some-other-ORG/"},
			"Pet[other_repo_name/types_A, B]", // second is local
			"repo_name/types_Pet[other_repo_name/types_A, repo_name/types_B]",
		},
	} {
		StripPackagePrefixes = tc.stripPrefixes
		var name string

		UsePackageName = false
		name = makeName(Mock{name: tc.name, str: tc.str, pkg: tc.pkg})
		assert.Equal(t, tc.expectedName, name)

		UsePackageName = true
		name = makeName(Mock{name: tc.name, str: tc.str, pkg: tc.pkg})
		assert.Equal(t, tc.expectedNameWithPackage, name)
	}

}
