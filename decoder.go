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
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"math/bits"
	"reflect"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Decoder reads and decodes data from Kubernets Object's metatdata
type Decoder struct {
	in     *metav1.ObjectMeta
	values []structField

	// indidacte if during Decode() the Decoder should generate ErrorList.
	// The ErrorList can be obtained by GetErrorList() method called on returned error.
	// Aditionally when True, the Decoder will not stop on first encountered deserialization error,
	// but instead will traverse fields to gather all possible decoding errors.
	AccumulateFieldErrors bool

	fieldErrors field.ErrorList
}

func assignToBool(out reflect.Value, in string) error {
	v, err := strconv.ParseBool(in)
	if err == nil {
		out.SetBool(v)
	}
	return err
}

func assignToInt(out reflect.Value, in string, bits int) error {
	v, err := strconv.ParseInt(in, 10, bits)
	if err == nil {
		out.SetInt(v)
	}
	return err
}

func assignToUInt(out reflect.Value, in string, bits int) error {
	v, err := strconv.ParseUint(in, 10, bits)
	if err == nil {
		out.SetUint(v)
	}
	return err
}

func assignToFloat(out reflect.Value, in string, bits int) error {
	v, err := strconv.ParseFloat(in, bits)
	if err == nil {
		out.SetFloat(v)
	}
	return err
}

func assignToArray(out reflect.Value, in string) error {
	values := strings.Split(in, itemSeparator)
	if out.Len() != len(values) {
		return errors.New("array elements number do not match")
	}
	for i, value := range values {
		if err := decodeUndefined(out.Index(i), value); err != nil {
			return fmt.Errorf("unable to decode array index %d, value: '%s': [%w]", i, value, err)
		}
	}
	return nil
}

func assignToSlice(out reflect.Value, in string) error {
	values := strings.Split(in, itemSeparator)
	slice := reflect.MakeSlice(out.Type(), len(values), len(values))
	for i, value := range values {
		if err := decodeUndefined(slice.Index(i), value); err != nil {
			return fmt.Errorf("unable to decode slice index %d, value: '%s': [%w]", i, value, err)
		}
	}
	out.Set(slice)
	return nil
}

func assignToMap(out reflect.Value, in string) error {
	values := strings.Split(in, itemSeparator)
	mp := reflect.MakeMapWithSize(out.Type(), len(values))
	for _, value := range values {
		elem := strings.Split(value, keyValueSeparator)
		if len(elem) != 2 {
			return fmt.Errorf("invalid map item syntax, expected <key>:<value>, got: %s", value)
		}
		// TODO? How to get existing value preallocated in map instead of creating new one?
		value := reflect.New(mp.Type().Elem()).Elem()
		if err := decodeUndefined(value, elem[1]); err != nil {
			return fmt.Errorf("unable to decode map item (key '%s', value: '%s'): [%w]", elem[0], elem[1], err)
		}
		mp.SetMapIndex(reflect.ValueOf(elem[0]), value)
	}
	out.Set(mp)
	return nil
}

func assignToPointer(out reflect.Value, in string) error {
	var realValue reflect.Value
	if out.IsZero() {
		realValue = reflect.New(out.Type().Elem())
	} else {
		realValue = out
	}
	if err := decodePrimitive(realValue.Elem(), in); err != nil {
		return fmt.Errorf("cannot assign value to pointer: [%w]", err)
	}
	out.Set(realValue)
	return nil
}

func decodePrimitive(out reflect.Value, in string) error {
	switch out.Kind() {
	case reflect.Bool:
		return assignToBool(out, in)
	case reflect.Int:
		return assignToInt(out, in, bits.UintSize)
	case reflect.Int8:
		return assignToInt(out, in, 8)
	case reflect.Int16:
		return assignToInt(out, in, 16)
	case reflect.Int32:
		return assignToInt(out, in, 32)
	case reflect.Int64:
		return assignToInt(out, in, 64)
	case reflect.Uint:
		return assignToUInt(out, in, bits.UintSize)
	case reflect.Uint8:
		return assignToUInt(out, in, 8)
	case reflect.Uint16:
		return assignToUInt(out, in, 16)
	case reflect.Uint32:
		return assignToUInt(out, in, 32)
	case reflect.Uint64:
		return assignToUInt(out, in, 64)
	case reflect.Float32:
		return assignToFloat(out, in, 32)
	case reflect.Float64:
		return assignToFloat(out, in, 64)
	case reflect.Array:
		return assignToArray(out, in)
	case reflect.Map:
		return assignToMap(out, in)
	case reflect.Pointer:
		return assignToPointer(out, in)
	case reflect.Slice:
		return assignToSlice(out, in)
	case reflect.String:
		out.SetString(in)
	default:
		return errors.New("unsupported type")
	}
	return nil
}

func decodeUsingTextUnmarshaler(out reflect.Value, in string) error {
	var fun reflect.Value

	if out.Kind() == reflect.Pointer && out.IsNil() {
		out.Set(reflect.New(out.Type().Elem()))
	}

	fun = findMethodByName(out, "UnmarshalText")
	if !fun.IsValid() || fun.IsZero() {
		return fmt.Errorf("type '%s' nor '*%s' doesn't implement encoding.TextUnmarshaler", out.Type().Name(), out.Type().Name())
	}
	ret := fun.Call([]reflect.Value{reflect.ValueOf([]byte(in))})
	if len(ret) != 1 {
		return fmt.Errorf("expected single return value, got %d", len(ret))
	}
	if err, ok := ret[0].Interface().(error); ok {
		return fmt.Errorf("failed to deserialize with encoding.TextUnmarshaler interface: [%w]", err)
	}
	return nil
}

func decodeUndefined(out reflect.Value, in string) error {
	if !out.IsValid() {
		return errors.New("unable to decode to invalid value")
	}
	// first try to check if TextUnmarshaler is defined for type
	if implements[encoding.TextUnmarshaler](out) {
		return decodeUsingTextUnmarshaler(out, in)
	}
	return decodePrimitive(out, in)
}

func decodeJson(out reflect.Value, in string) error {
	if out.Kind() == reflect.Pointer && out.IsNil() {
		out.Set(reflect.New(out.Type().Elem()))
	} else {
		out = out.Addr()
	}
	return json.Unmarshal([]byte(in), out.Interface())
}

func decodeCustom(out reflect.Value, meta *metav1.ObjectMeta) error {
	var fun reflect.Value

	if out.Kind() == reflect.Pointer && out.IsNil() {
		out.Set(reflect.New(out.Type().Elem()))
	}

	fun = findMethodByName(out, "UnmarshalFromMetadata")
	if !fun.IsValid() || fun.IsZero() {
		return fmt.Errorf("type '%s' nor '*%s' doesn't implement metaser.MetadataUnmarshaler", out.Type().Name(), out.Type().Name())
	}
	ret := fun.Call([]reflect.Value{reflect.ValueOf(meta)})
	if len(ret) != 1 {
		return fmt.Errorf("expected single return value, got %d", len(ret))
	}
	if err, ok := ret[0].Interface().(error); ok {
		return fmt.Errorf("failed to deserialize with metaser.MetadataUnmarshaler interface: [%w]", err)
	}
	return nil
}

func decode(out reflect.Value, in string, enc encoder) error {
	switch enc {
	case encoder(undefined):
		return decodeUndefined(out, in)
	case jsonEnc:
		return decodeJson(out, in)
	}
	return nil
}

func (dec *Decoder) decodeField(dv *structField) error {
	tag, err := parseTag(dv.tag)
	if err != nil {
		return fmt.Errorf("unable to parse tag '%s': [%w]", string(dv.tag), err)
	}
	if tag == nil || tag.dir == out {
		return nil
	}

	if tag.inline {
		dec.values = appendValues(dec.values, dv.value)
		return nil
	}

	switch tag.source {
	case name:
		err = decodePrimitive(dv.value, dec.in.Name)
	case namespace:
		err = decodePrimitive(dv.value, dec.in.Namespace)
	case label:
		if val, ok := dec.in.Labels[tag.value]; ok {
			err = decode(dv.value, val, tag.enc)
		}
	case annotation:
		if val, ok := dec.in.Annotations[tag.value]; ok {
			err = decode(dv.value, val, tag.enc)
		}
	case custom:
		err = decodeCustom(dv.value, dec.in)
	}

	if err != nil && dec.AccumulateFieldErrors {
		dec.fieldErrors = append(dec.fieldErrors, field.TypeInvalid(field.NewPath("metadata").Child(tag.source.String()),
			tag.value, err.Error()))
		return nil
	}

	if err != nil {
		return fmt.Errorf("%s %s: [%w]", tag.source, tag.value, err)
	}

	return nil
}

// Decode reads data from K8s object metadata and stores them in v.
//
// See package documentation for details about deserialization.
func (dec *Decoder) Decode(v any) error {
	dec.values = appendValues(dec.values, reflect.ValueOf(v))
	dec.fieldErrors = field.ErrorList{}

	for {
		if len(dec.values) == 0 {
			break
		}
		v := dec.values[len(dec.values)-1]
		dec.values = dec.values[:len(dec.values)-1]

		if err := dec.decodeField(&v); err != nil {
			return fmt.Errorf("unable to decode value: [%w]", err)
		}
	}

	if len(dec.fieldErrors) > 0 {
		return &decodeError{message: "unable to decode value", fieldErrors: dec.fieldErrors}
	}

	return nil
}

// NewDecoder returns new Decoder that reads data from meta.
func NewDecoder(meta *metav1.ObjectMeta) *Decoder {
	return &Decoder{
		in: meta,
	}
}

// Unmarshal reads data from K8s object metadata and stores them in v.
func Unmarshal(meta *metav1.ObjectMeta, v any) error {
	return NewDecoder(meta).Decode(v)
}
