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
	"reflect"
)

func dereference(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Pointer {
		return v.Elem()
	}
	return v
}

func method(v reflect.Value, name string) reflect.Value {
	fun := v.MethodByName(name)
	if !fun.IsValid() {
		fun = v.Addr().MethodByName(name)
	}
	return fun
}

func implements[T any](out reflect.Value) bool {
	return out.Type().Implements(reflect.TypeOf((*T)(nil)).Elem()) ||
		(out.CanAddr() && out.Addr().Type().Implements(reflect.TypeOf((*T)(nil)).Elem()))
}

type tp struct {
	t reflect.Type
	p []int
}

func visit(root reflect.Type, visitor func(reflect.Type, []int) (bool, error)) error {
	q := []tp{{root, nil}}
	visited := map[reflect.Type]struct{}{}

	for l := len(q); l > 0; l = len(q) {
		e := q[l-1]
		q = q[:l-1]
		if _, ok := visited[e.t]; ok {
			continue
		}
		visited[e.t] = struct{}{}
		expand, err := visitor(e.t, e.p)
		if err != nil {
			return err
		}
		if !expand {
			continue
		}
		if e.t.Kind() == reflect.Struct {
			for i := 0; i < e.t.NumField(); i++ {
				p := make([]int, len(e.p)+1)
				copy(p, append(e.p, i))
				q = append(q, tp{e.t.Field(i).Type, p})
			}
		}
		if e.t.Kind() == reflect.Pointer {
			q = append(q, tp{e.t.Elem(), e.p})
		}
	}
	return nil
}
