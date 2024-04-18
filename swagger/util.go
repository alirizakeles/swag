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
	"reflect"
	"strings"
)

// UsePackageName can be set to true to add package prefix of generated definition names
var UsePackageName = false

// StripPackagePrefixes can be set to remove leading strings from long package names, eg github.com/some-ORG/
// So github.com/some-ORG/repo/types.Pet becomes repo/types.Pet
var StripPackagePrefixes []string

func makeRef(name string) string {
	return fmt.Sprintf("#/definitions/%s", url.QueryEscape(name))
}

type parsedType interface {
	fmt.Stringer

	// handlePackageName handles all transformations on package names.
	//
	// If UsePackageName is true, any non-builtin type with an empty package will be set to use `packageName`; if UsePackageName is false, the reverse transformation is applied.
	// Any package which matches any entry in StripPackagePrefixes will have that prefix stripped.
	// All generic arguments will be transformed recursively.
	handlePackageName(string)
}

var _ parsedType = &parsedNamed{}

type parsedNamed struct {
	pkg     string
	name    string
	generic []parsedType
}

func (p parsedNamed) String() string {
	s := p.name

	if p.pkg != "" {
		s = p.pkg + "." + s
	}

	if p.generic != nil {
		sep := "["
		for _, g := range p.generic {
			s += sep + g.String()
			sep = ", "
		}
		s += "]"
	}

	return s
}
func (ty parsedNamed) isBuiltin() bool {
	for k := reflect.Invalid; k <= reflect.UnsafePointer; k++ {
		if k.String() == ty.name {
			return true
		}
	}
	return false
}
func (ty *parsedNamed) handlePackageName(packageName string) {
	if UsePackageName {
		if ty.pkg == "" && !ty.isBuiltin() {
			ty.pkg = packageName
		}
	} else {
		if ty.pkg == packageName {
			ty.pkg = ""
		}
	}

	for _, pfx := range StripPackagePrefixes {
		if strings.HasPrefix(ty.pkg, pfx) {
			ty.pkg = strings.TrimPrefix(ty.pkg, pfx)
			break
		}
	}

	for _, g := range ty.generic {
		g.handlePackageName(packageName)
	}
}

var _ parsedType = &parsedMap{}

type parsedMap struct {
	key   parsedType
	value parsedType
}

func (ty parsedMap) String() string {
	return fmt.Sprintf("map_%s_to_%s", ty.key, ty.value)
}
func (ty *parsedMap) handlePackageName(packageName string) {
	ty.key.handlePackageName(packageName)
	ty.value.handlePackageName(packageName)
}

var _ parsedType = &parsedSlice{}

type parsedSlice struct {
	count string
	elem  parsedType
}

func (ty parsedSlice) String() string {
	if ty.count != "" {
		return fmt.Sprintf("arr_%s_%s", ty.count, ty.elem)
	} else {
		return fmt.Sprintf("arr_%s", ty.elem)
	}
}
func (ty *parsedSlice) handlePackageName(packageName string) {
	ty.elem.handlePackageName(packageName)
}

func parseArrayCount(input string) (string, string) {
	var s string

	for len(input) > 0 && '0' <= input[0] && input[0] <= '9' {
		s += string(input[0])
		input = input[1:]
	}

	return s, input
}

var _ parsedType = &parsedPtr{}

type parsedPtr struct {
	elem parsedType
}

func (ty parsedPtr) String() string {
	return fmt.Sprintf("ptr_%s", ty.elem)
}
func (ty *parsedPtr) handlePackageName(packageName string) {
	ty.elem.handlePackageName(packageName)
}

// parseType parses a type into a parsedType.
//
//	Foo => parsedNamed{name: "Foo"}
//	my.Foo => parsedNamed{pkg: "my", name: "Foo"}
//	Foo[my.Bar] => parsedNamed{name: "Foo", generic: [parsedNamed{pkg: "my", name: "Bar"}]
//	Foo[my.Bar[Baz]] => parsedNamed{name: "Foo", generic: [parsedNamed{pkg: "my", name: "Bar", generic: [{name: "Baz"}]}]
//	[]Foo => parsedSlice{ty: {name: "Foo"}}
//	map[Foo]Bar => parsedMap{key: {name: "Foo"}, value: {name: "Bar"}}
func parseType(s string) (parsedType, string) {
	pkg := ""
	name := ""
	var generic []parsedType

	if strings.HasPrefix(s, "map[") {
		key, rest := parseType(s[4:])
		if rest[0] != ']' {
			panic(fmt.Sprintf("failed to parse type %q: bad map", s))
		}
		value, rest := parseType(rest[1:])

		return &parsedMap{
			key:   key,
			value: value,
		}, rest
	}

	if strings.HasPrefix(s, "[") {
		count, rest := parseArrayCount(s[1:])
		if rest[0] != ']' {
			panic(fmt.Sprintf("failed to parse type %q: bad array/slice", s))
		}

		elem, rest := parseType(rest[1:])
		return &parsedSlice{
			count: count,
			elem:  elem,
		}, rest
	}

	if strings.HasPrefix(s, "*") {
		elem, rest := parseType(s[1:])
		return &parsedPtr{
			elem: elem,
		}, rest
	}

loop:
	for len(s) != 0 {
		switch s[0] {
		case '.':
			pkg += name + "."
			name = ""
			s = s[1:]
		case '/':
			pkg += name + "/"
			name = ""
			s = s[1:]
		case '[':
			s = s[1:]
			for len(s) != 0 {
				this, rest := parseType(s)
				generic = append(generic, this)
				switch rest[0] {
				case ',':
					s = rest[1:]
					if rest[1] == ' ' {
						s = rest[2:]
					}
				case ']':
					s = rest[1:]
					break loop
				}
			}
		case ',', ']':
			break loop
		default:
			name += string(s[0])
			s = s[1:]
		}
	}

	pkg = strings.TrimSuffix(pkg, ".")

	return &parsedNamed{
		pkg:     pkg,
		name:    name,
		generic: generic,
	}, s
}

type reflectType interface {
	PkgPath() string
	Name() string
	String() string
}

func makeName(t reflect.Type) string {
	ty := reflectParseType(t)
	name := ty.String()

	name = strings.TrimSpace(name)
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, "/", "__", -1) // slashes are problematic due to the name ending up in the fragment of a URL parse

	return name
}

func reflectParseType(t reflect.Type) parsedType {
	if t.Name() != "" {
		p, rest := parseType(t.Name())
		if rest != "" {
			panic(fmt.Sprintf("failed to parse type %q, rest=%q", t.Name(), rest))
		}
		p.handlePackageName(t.PkgPath())
		return p
	}
	switch t.Kind() {
	case reflect.Array:
		return &parsedSlice{
			count: fmt.Sprintf("%d", t.Len()),
			elem:  reflectParseType(t.Elem()),
		}
	case reflect.Slice:
		return &parsedSlice{
			elem: reflectParseType(t.Elem()),
		}
	case reflect.Map:
		return &parsedMap{
			key:   reflectParseType(t.Key()),
			value: reflectParseType(t.Elem()),
		}
	case reflect.Ptr:
		return &parsedPtr{
			elem: reflectParseType(t.Elem()),
		}
	default:
		// hopefully only builtins make it here; if we have to call `t.String()`, we don't get full package information
		p, rest := parseType(t.String())
		if rest != "" {
			panic(fmt.Sprintf("failed to parse type %q, rest=%q", t.String(), rest))
		}
		return p
	}
}
