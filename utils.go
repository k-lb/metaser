/*
Copyright (c) 2023 Samsung Electronics Co., Ltd All Rights Reserved

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

func findMethodByName(v reflect.Value, name string) reflect.Value {
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

func appendValues(values []structField, v reflect.Value) []structField {
	v = dereference(v)

	if v.Kind() != reflect.Struct {
		return values
	}

	for i := 0; i < v.NumField(); i++ {
		values = append(values, structField{
			value: v.Field(i),
			tag:   v.Type().Field(i).Tag,
		})
	}
	return values
}
