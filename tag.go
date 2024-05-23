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

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type parsedTag struct {
	source    source
	enc       encoder
	dir       dir
	value     string
	inline    bool
	omitempty bool
	immutable bool
	aliases   []string
	setOnce   bool
}

func parseEncoding(expr string) (encoder, error) {
	switch expr {
	case jsonKey:
		return encoder(jsonEnc), nil
	case customKey:
		return encoder(custom), nil
	case "":
		return encoder(undefined), nil
	default:
		return encoder(undefined), errors.New("unsupported type")
	}
}

// parseTag returns parsed k8s tag or nil if tag is not defined for struct field.
func parseTag(tag reflect.StructTag) (pt *parsedTag, err error) {

	pt = &parsedTag{
		source:    source(undefined),
		enc:       encoder(undefined),
		dir:       dir(inout),
		value:     "",
		inline:    false,
		omitempty: false,
		immutable: false,
	}

	k8sTag := ""
	for _, f := range strings.Fields(string(tag)) {
		if strings.HasPrefix(f, k8sKey) {
			k8sTag = strings.TrimSuffix(strings.TrimPrefix(f, k8sKey+`:"`), `"`)
			break
		}
	}

	if k8sTag == "" {
		return nil, nil
	}

	for _, f := range strings.Split(k8sTag, ",") {
		switch f {
		case nameKey:
			pt.source = name
		case namespaceKey:
			pt.source = namespace
		case inlineKey:
			pt.inline = true
		case inKey:
			pt.dir = in
		case outKey:
			pt.dir = out
		case inoutKey:
			pt.dir = inout
		case omitEmptyKey:
			pt.omitempty = true
		case immutableKey:
			pt.immutable = true
		case setOnceKey:
			pt.setOnce = true
		default:
			// handle key:value pairs
			keyvals := strings.Split(f, ":")
			if len(keyvals) != 2 {
				return nil, fmt.Errorf("invalid encoding tag syntax. Unknown k8s option: '%s'", f)
			}
			switch keyvals[0] {
			case encodingKey:
				if pt.enc, err = parseEncoding(keyvals[1]); err != nil {
					return nil, fmt.Errorf("invalid encoding value. Expected one of [json], got '%s': [%w]", keyvals[1], err)
				}
			case annotationKey:
				pt.source = annotation
				pt.value = keyvals[1]
			case labelKey:
				pt.source = label
				pt.value = keyvals[1]
			case aliasesKey:
				pt.aliases = strings.Split(keyvals[1], ";")
			default:
				return nil, fmt.Errorf("invalid tag syntax. Expected <option>:<value>, unknown option: '%s'", keyvals[0])
			}
		}
	}
	return pt, nil
}
