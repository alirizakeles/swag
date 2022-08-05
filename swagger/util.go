// Package swagger ...
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
//
package swagger

import (
	"fmt"
	"path/filepath"
	"strings"
)

// UsePackageName can be set to true to add package prefix of generated definition names
var UsePackageName = false

func makeRef(name string) string {
	return fmt.Sprintf("#/definitions/%v", name)
}

type reflectType interface {
	PkgPath() string
	Name() string
	String() string
}

func makeName(t reflectType) string {
	name := t.Name()
	if name != "" && t.PkgPath() != "" && UsePackageName {
		name = filepath.Base(t.PkgPath()) + name
	} else if name != "" {
	} else {
		name = t.String()
		name = strings.ReplaceAll(name, "[]", "arr_")
		name = strings.ReplaceAll(name, "*", "ptr_")
		name = strings.ReplaceAll(name, "[", "_")
		name = strings.ReplaceAll(name, "]", "_to_")
	}
	return strings.Replace(name, "-", "_", -1)
}
