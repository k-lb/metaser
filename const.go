/*
Copyright (c) 2023 - 2024 Samsung Electronics Co., Ltd All Rights Reserved

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metaser

const (
	k8sKey            = "k8s"
	nameKey           = "name"
	namespaceKey      = "namespace"
	dataKey           = "data"
	annotationKey     = "annotation"
	labelKey          = "label"
	inKey             = "in"
	outKey            = "out"
	inoutKey          = "inout"
	encodingKey       = "enc"
	jsonKey           = "json"
	customKey         = "custom"
	inlineKey         = "inline"
	itemSeparator     = ","
	keyValueSeparator = ":"
	omitEmptyKey      = "omitempty"
	immutableKey      = "immutable"
	aliasesKey        = "aliases"
	setOnceKey        = "setonce"
)

type source int
type encoder int
type dir int

const undefined int = 0

const (
	name source = iota + 1
	namespace
	annotation
	label
)

const (
	in dir = iota + 1
	out
	inout
)

const (
	jsonEnc encoder = iota + 1
	custom
)

func (s source) String() string {
	switch s {
	case name:
		return nameKey
	case namespace:
		return namespaceKey
	case annotation:
		return annotationKey
	case label:
		return labelKey
	}
	return "undefined source"
}
