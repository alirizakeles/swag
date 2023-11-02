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
package swagger

import (
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

// UsePackageName can be set to true to add package prefix of generated definition names
var UsePackageName = false

// StripPackagePrefixes can be set to remove leading strings from long package names, eg github.com/some-ORG/
// So github.com/some-ORG/repo/types.Pet becomes repo/types.Pet
var StripPackagePrefixes []string

var genericTypeRegex = regexp.MustCompile(`(?P<type>[\w.\-/]+)\[(?P<typeParams>[\w.\-/,\s]+)\]`)

func makeRef(name string) string {
	return fmt.Sprintf("#/definitions/%s", url.QueryEscape(name))
}

type reflectType interface {
	PkgPath() string
	Name() string
	String() string
}

func makeName(t reflectType) string {
	name := t.Name()

	matches := genericTypeRegex.FindStringSubmatch(name)
	if len(matches) > 1 { // handle generic type names
		return handleGenericTypeNames(matches, t.PkgPath())
	}

	if name != "" {
		name = prefixPackageName(name, t.PkgPath())
	} else {
		name = t.String()
		name = strings.ReplaceAll(name, "[]", "arr_")
		name = strings.ReplaceAll(name, "*", "ptr_")
		name = strings.ReplaceAll(name, "[", "_")
		name = strings.ReplaceAll(name, "]", "_to_")
	}
	return formatName(name)
}

// handleGenericTypeNames generates shorter Generic types
// types.A[types.B, types.C] => A[B,C]
func handleGenericTypeNames(matches []string, packageName string) string {
	var genericName, typeParamNames string

	typeIndex := genericTypeRegex.SubexpIndex("type")
	if typeIndex > -1 {
		genericName = prefixPackageName(matches[typeIndex], packageName)
		genericName = formatName(genericName)
	}

	paramIndex := genericTypeRegex.SubexpIndex("typeParams")
	if typeIndex > -1 {
		typeParamNames = matches[paramIndex]
	}

	var cleanTypeParamNames []string
	for _, typeParamsName := range strings.Split(typeParamNames, ",") {
		typeParamsName = prefixPackageName(typeParamsName, packageName)
		typeParamsName = formatName(typeParamsName)

		cleanTypeParamNames = append(cleanTypeParamNames, formatName(typeParamsName))
	}
	return fmt.Sprintf("%s[%s]", genericName, strings.Join(cleanTypeParamNames, ", "))
}

// Given
//
//	StripPackagePrefixes = []string{"gitlab.com/some-ORG/"}
//
// Then
//
//	gitlab.com/some-ORG/repo-name/types.A => repo_name/types_A
//	gitlab.com/other-ORG/repo-name/types.A => gitlab_com/other_ORG/repo_name/types_A
func formatName(name string) string {
	name = strings.TrimSpace(name)
	for _, strip := range StripPackagePrefixes {
		name = strings.TrimPrefix(name, strip)
	}
	name = strings.Replace(name, ".", "_", -1)
	return strings.Replace(name, "-", "_", -1)
}

// Given
//
//	packageName = types
//	UsePackageName = false
//
// Then
//
//	types.Pet => Pet
//	Pet => Pet
//	other.Pet => other.Pet
//
// Given
//
//	UsePackageName true
//
// Then
//
//	types.Pet => types.Pet
//	Pet => types.Pet
//	other.Pet => other.Pet
func prefixPackageName(name, packageName string) string {
	name = strings.TrimSpace(name)
	if isBuiltinType(name) {
		return name
	}
	basename := filepath.Base(name)
	alreadyHasPrefix := strings.HasPrefix(name, packageName)
	plainType := basename == name

	if !UsePackageName {
		if !alreadyHasPrefix && !plainType {
			return name
		}
		ss := strings.Split(basename, ".")
		return ss[len(ss)-1]
	}

	if alreadyHasPrefix {
		return name
	}

	if packageName != "" && plainType {
		name = packageName + "." + name
	}

	return name
}

func isBuiltinType(s string) bool {
	for k := reflect.Invalid; k <= reflect.UnsafePointer; k++ {
		if k.String() == s {
			return true
		}
	}
	return false
}
