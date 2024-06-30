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
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Encoder encodes and writes data into Kubernets Object's metatdata
type Encoder struct{}

// internal struct represents context of encoding operation.
type encodeContext struct {
	meta metav1.Object
	out  struct {
		Labels      map[string]string
		Annotations map[string]string
	}
	values []structField
}

// EncodeOption to be passed to Encode()
type EncodeOption func(enc *encodeContext)

func encodeUsingTextMarshaler(in reflect.Value) (string, error) {
	fun := method(in, "MarshalText")
	if !fun.IsValid() || fun.IsZero() {
		return "", fmt.Errorf("type '%s' or '*%s' doesn't implement encoding.TextMarshaler interface", in.Type().Name(), in.Type().Name())
	}
	ret := fun.Call([]reflect.Value{})
	if len(ret) != 2 {
		return "", fmt.Errorf("expected two return values, got %d", len(ret))
	}
	if err, ok := ret[1].Interface().(error); ok {
		return "", fmt.Errorf("failed to serialize with encoding.TextMarshaler interface: [%w]", err)
	}
	if ret[0].Kind() != reflect.Slice && ret[0].Elem().Kind() != reflect.Uint8 {
		return "", fmt.Errorf("failed to serialize with encoding.TextMarshaler interface: return type is not byte slice")
	}
	return string(ret[0].Bytes()), nil
}

func encodeUndefined(in reflect.Value) (string, error) {
	if !in.IsValid() {
		return "", fmt.Errorf("unable to encode invalid value")
	}
	// first try to check if TextMarshaler is defined for type
	if implements[encoding.TextMarshaler](in) {
		return encodeUsingTextMarshaler(in)
	}
	if isOption(in) {
		return encodeOption(in)
	}
	return encodePrimitive(in)
}

func encodeOption(in reflect.Value) (string, error) {
	isSome := in.Field(isSetFieldIndex)
	if isSome.Bool() {
		return encodeUndefined(in.Field(valueFieldIndex))
	}
	return "", nil
}

func encodeJson(in reflect.Value) (string, error) {
	val, err := json.Marshal(in.Interface())
	if err != nil {
		return "", fmt.Errorf("cannot marshal value: [%w]", err)
	}
	return string(val), nil
}

func assignBool(in reflect.Value, out *string) error {
	*out = strconv.FormatBool(in.Bool())
	return nil
}

func assignInt(in reflect.Value, out *string) error {
	*out = strconv.FormatInt(in.Int(), 10)
	return nil
}

func assignUInt(in reflect.Value, out *string) error {
	*out = strconv.FormatUint(in.Uint(), 10)
	return nil
}

func assignFloat(in reflect.Value, out *string, bitSize int) error {
	*out = strconv.FormatFloat(in.Float(), 'f', -1, bitSize)
	return nil
}

func assignArray(in reflect.Value, out *string) error {
	elems := make([]string, in.Len())
	for i := 0; i < in.Len(); i++ {
		v, err := encodeUndefined(in.Index(i))
		if err != nil {
			return fmt.Errorf("cannot encode array element at index %d: [%w]", i, err)
		}
		elems[i] = v
	}
	*out = strings.Join(elems, itemSeparator)
	return nil
}

func assignMap(in reflect.Value, out *string) error {
	elems := make([]string, in.Len())
	iter := in.MapRange()
	i := 0
	for iter.Next() {
		v := iter.Value()
		k := iter.Key()
		ev, err := encodeUndefined(v)
		if err != nil {
			return fmt.Errorf("cannot encode map value element: [%w]", err)
		}
		ek, err := encodeUndefined(k)
		if err != nil {
			return fmt.Errorf("cannot encode map key element: [%w]", err)
		}
		elems[i] = strings.Join([]string{ek, ev}, keyValueSeparator)
		i++
	}
	*out = strings.Join(elems, itemSeparator)
	return nil
}

func assignPointer(in reflect.Value, out *string) error {
	if in.IsNil() {
		return nil
	}
	v, err := encodeUndefined(in.Elem())
	if err != nil {
		return fmt.Errorf("cannot encode pointer: [%w]", err)
	}
	*out = v
	return nil
}

func assignSlice(in reflect.Value, out *string) error {
	return assignArray(in, out)
}

func encodePrimitive(in reflect.Value) (out string, err error) {
	switch in.Kind() {
	case reflect.Bool:
		err = assignBool(in, &out)
	case reflect.Int:
		err = assignInt(in, &out)
	case reflect.Int8:
		err = assignInt(in, &out)
	case reflect.Int16:
		err = assignInt(in, &out)
	case reflect.Int32:
		err = assignInt(in, &out)
	case reflect.Int64:
		err = assignInt(in, &out)
	case reflect.Uint:
		err = assignUInt(in, &out)
	case reflect.Uint8:
		err = assignUInt(in, &out)
	case reflect.Uint16:
		err = assignUInt(in, &out)
	case reflect.Uint32:
		err = assignUInt(in, &out)
	case reflect.Uint64:
		err = assignUInt(in, &out)
	case reflect.Float32:
		err = assignFloat(in, &out, 32)
	case reflect.Float64:
		err = assignFloat(in, &out, 64)
	case reflect.Array:
		err = assignArray(in, &out)
	case reflect.Map:
		err = assignMap(in, &out)
	case reflect.Pointer:
		err = assignPointer(in, &out)
	case reflect.Slice:
		err = assignSlice(in, &out)
	case reflect.String:
		out = in.String()
		err = nil
	default:
		return "", fmt.Errorf("unsupported type")
	}
	return out, err
}

func encode(in reflect.Value, enc encoder, meta metav1.Object) (string, error) {
	switch enc {
	case encoder(undefined):
		return encodeUndefined(in)
	case jsonEnc:
		return encodeJson(in)
	case custom:
		return "", encodeCustom(in, meta)
	default:
		return "", fmt.Errorf("unsupported encoding")
	}
}

func encodeCustom(out reflect.Value, meta metav1.Object) error {
	var fun reflect.Value

	if out.Kind() == reflect.Pointer && out.IsNil() {
		return nil
	}

	fun = method(out, "MarshalToMetadata")
	if !fun.IsValid() || fun.IsZero() {
		return fmt.Errorf("type '%s' or '*%s' doesn't implement metaser.MetadataMarshaler interface", out.Type().Name(), out.Type().Name())
	}
	ret := fun.Call([]reflect.Value{reflect.ValueOf(meta)})
	if len(ret) != 1 {
		return fmt.Errorf("expected single return value, got %d", len(ret))
	}
	if err, ok := ret[0].Interface().(error); ok {
		return fmt.Errorf("failed to serialize with metaser.MetadataMarshaler interface: [%w]", err)
	}
	return nil
}

func encodeField(ec *encodeContext, dv *structField) error {
	var val string
	var err error

	// do not encoded fields that are marked as 'input only' or 'inline'.
	// fields within inline field will be encoded by separate calls to encodeField.
	if dv.tag == nil || dv.tag.dir == in || dv.tag.inline {
		return nil
	}

	if dv.tag.omitempty && dv.value.IsZero() {
		switch dv.tag.source {
		case label:
			delete(ec.out.Labels, dv.tag.value)
		case annotation:
			delete(ec.out.Annotations, dv.tag.value)
		}
		return nil
	}

	switch dv.tag.source {
	case name:
		if val, err = encodePrimitive(dv.value); err == nil {
			ec.meta.SetName(val)
		}
	case namespace:
		if val, err = encodePrimitive(dv.value); err == nil {
			ec.meta.SetNamespace(val)
		}
	case label:
		if val, err = encode(dv.value, dv.tag.enc, ec.meta); err == nil {
			ec.out.Labels[dv.tag.value] = val
		}
	case annotation:
		if val, err = encode(dv.value, dv.tag.enc, ec.meta); err == nil {
			ec.out.Annotations[dv.tag.value] = val
		}
	case source(undefined):
		_, err = encode(dv.value, dv.tag.enc, ec.meta)
	}

	return err
}

func appendFieldValues(values []structField, v reflect.Value) ([]structField, error) {
	v = dereference(v)

	if v.Kind() != reflect.Struct {
		return values, nil
	}

	for i := 0; i < v.NumField(); i++ {
		ptag, err := parseTag(v.Type().Field(i).Tag)
		if err != nil {
			return nil, err
		}
		values = append(values, structField{
			value: v.Field(i),
			tag:   ptag,
		})
	}
	return values, nil
}

// Encode reads data from v and writes it into K8s object metadata.
//
// See package documentation for details about serialization.
func (*Encoder) Encode(v any, meta metav1.Object, options ...EncodeOption) error {
	var err error
	value := reflect.ValueOf(v)

	if value.Kind() != reflect.Pointer {
		return fmt.Errorf("expected pointer to value")
	}

	ec := &encodeContext{
		meta: meta,
	}

	for _, opt := range options {
		opt(ec)
	}

	ec.out.Annotations = meta.GetAnnotations()
	if ec.out.Annotations == nil {
		ec.out.Annotations = map[string]string{}
		meta.SetAnnotations(ec.out.Annotations)
	}

	ec.out.Labels = meta.GetLabels()
	if ec.out.Labels == nil {
		ec.out.Labels = map[string]string{}
		meta.SetLabels(ec.out.Labels)
	}

	ec.values, err = appendFieldValues(ec.values, value)
	if err != nil {
		return err
	}

	for len(ec.values) > 0 {
		v := ec.values[len(ec.values)-1]
		ec.values = ec.values[:len(ec.values)-1]
		if err = encodeField(ec, &v); err != nil {
			return fmt.Errorf("unable to process value: [%w]", err)
		}

		if v.tag != nil && v.tag.inline {
			if ec.values, err = appendFieldValues(ec.values, v.value); err != nil {
				return err
			}
		}
	}

	return nil
}

// NewEncoder returns new Encoder that writes data into meta.
func NewEncoder() *Encoder {
	return &Encoder{}
}

// Marshal reads data from v and writes it into K8s object metadata using default Encoder.
func Marshal(v any, meta metav1.Object, options ...EncodeOption) error {
	return NewEncoder().Encode(v, meta, options...)
}
