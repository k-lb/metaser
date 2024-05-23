/*
Copyright (c) 2024 Samsung Electronics Co., Ltd All Rights Reserved

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
	"reflect"
)

type fieldInfo struct {
	path []int
	tag  parsedTag
}

type cache struct {
	CachedType             reflect.Type
	NameFastAccess         []fieldInfo
	NamespaceFastAccess    []fieldInfo
	AnnotationFastAccess   map[string][]fieldInfo
	LabelsFastAccess       map[string][]fieldInfo
	CustomFieldsFastAccess []fieldInfo
}

func (c *cache) build(root reflect.Value) error {
	if c.CachedType == root.Type() {
		return nil
	}

	c.AnnotationFastAccess = map[string][]fieldInfo{}
	c.LabelsFastAccess = map[string][]fieldInfo{}
	c.CustomFieldsFastAccess = nil
	c.NameFastAccess = nil
	c.NamespaceFastAccess = nil

	err := visit(root.Type(), func(t reflect.Type, path []int) (bool, error) {
		if t.Kind() == reflect.Pointer {
			return true, nil
		}
		if t.Kind() != reflect.Struct {
			return false, nil
		}
		recurse := false
		for i := 0; i < t.NumField(); i++ {
			pt, err := parseTag(t.Field(i).Tag)
			if err != nil {
				return false, err
			}
			if pt == nil {
				continue
			}
			recurse = true
			item := fieldInfo{append(path, i), *pt}
			switch pt.source {
			case name:
				c.NameFastAccess = append(c.NameFastAccess, item)
			case namespace:
				c.NamespaceFastAccess = append(c.NamespaceFastAccess, item)
			case annotation:
				v := c.AnnotationFastAccess[pt.value]
				v = append(v, item)
				c.AnnotationFastAccess[pt.value] = v
			case label:
				v := c.LabelsFastAccess[pt.value]
				v = append(v, item)
				c.LabelsFastAccess[pt.value] = v
			case source(undefined):
				if pt.enc == custom {
					c.CustomFieldsFastAccess = append(c.CustomFieldsFastAccess, item)
				}
			}
			recurse = recurse || pt.inline
		}
		return recurse, nil
	})
	if err == nil {
		c.CachedType = root.Type()
	}
	return err
}
