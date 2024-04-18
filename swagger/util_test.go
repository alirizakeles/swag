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
	"encoding/json"
	"reflect"
	"strings"
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

type G0 struct{}
type G1[A any] struct{}
type G2[A any, B any] struct{}

func TestMakeSchema(t *testing.T) {

	self := reflect.TypeOf(G0{}).PkgPath()
	self = strings.Replace(self, ".", "_", -1)
	self = strings.Replace(self, "/", "__", -1)

	self_no_gh := strings.TrimPrefix(self, "github_com__")

	for _, tc := range []struct {
		ty                      interface{}
		stripPrefixes           []string
		expectedName            string
		expectedNameWithPackage string
	}{
		{
			G0{},
			[]string{},
			"G0",
			self + "_G0",
		},
		{
			"",
			[]string{},
			"string",
			"string",
		},
		{
			int32(0),
			[]string{},
			"int32",
			"int32",
		},
		{
			G1[string]{},
			[]string{},
			"G1[string]",
			self + "_G1[string]",
		},
		{
			G1[uint64]{},
			[]string{"github.com/"},
			"G1[uint64]",
			self_no_gh + "_G1[uint64]",
		},
		{
			[]G0{},
			[]string{},
			"arr_G0",
			"arr_" + self + "_G0",
		},
		{
			[4]G0{G0{}, G0{}, G0{}, G0{}},
			[]string{},
			"arr_4_G0",
			"arr_4_" + self + "_G0",
		},
		{
			json.RawMessage{},
			[]string{},
			"RawMessage",
			"encoding__json_RawMessage",
		},
		{
			[]G1[json.RawMessage]{},
			[]string{},
			"arr_G1[encoding__json_RawMessage]",
			"arr_" + self + "_G1[encoding__json_RawMessage]",
		},
		{
			[]G1[json.RawMessage]{},
			[]string{"encoding/"},
			"arr_G1[json_RawMessage]",
			"arr_" + self + "_G1[json_RawMessage]",
		},
		{
			[]json.RawMessage{},
			[]string{},
			"arr_RawMessage",
			"arr_encoding__json_RawMessage",
		},
		{
			map[G0]json.RawMessage{},
			[]string{},
			"map_G0_to_RawMessage",
			"map_" + self + "_G0_to_encoding__json_RawMessage",
		},
		{
			&G0{},
			[]string{},
			"ptr_G0",
			"ptr_" + self + "_G0",
		},
		{
			G1[G0]{},
			[]string{},
			"G1[G0]",
			self + "_G1[" + self + "_G0]",
		},
		{
			G2[G0, G0]{},
			[]string{},
			"G2[G0, G0]",
			self + "_G2[" + self + "_G0, " + self + "_G0]",
		},
		{
			G1[G1[G0]]{},
			[]string{},
			"G1[G1[G0]]",
			self + "_G1[" + self + "_G1[" + self + "_G0]]",
		},
		{
			G1[G1[G0]]{},
			[]string{"github.com/"},
			"G1[G1[G0]]",
			self_no_gh + "_G1[" + self_no_gh + "_G1[" + self_no_gh + "_G0]]",
		},
	} {
		StripPackagePrefixes = tc.stripPrefixes
		var name string

		rty := reflect.TypeOf(tc.ty)

		UsePackageName = false
		name = makeName(rty)
		assert.Equal(t, tc.expectedName, name)

		UsePackageName = true
		name = makeName(rty)
		assert.Equal(t, tc.expectedNameWithPackage, name)
	}

}
