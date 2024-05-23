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
	"errors"
	"fmt"
	"math/bits"
	"reflect"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Decoder reads and decodes data from Kubernets Resource metatdata
type Decoder struct {
	meta        *metav1.ObjectMeta
	fieldErrors field.ErrorList
	cache       cache
	root        reflect.Value

	performValidation     bool
	accumulateFieldErrors bool
	skipDefaultWorkload   bool
	filter                fieldFilter
}

// DecodeOption to be passed to Decode()
type DecodeOption func(dec *Decoder)

type fieldFilter func(field *fieldInfo) bool

func (f fieldFilter) Apply(info *fieldInfo) bool {
	if f != nil {
		return f(info)
	}
	return true
}

// AccumulateFieldErrors enforces decoder to accumulate
// all encountered decode errors instead of aborting on first found one.
// the list of errors can be obtained with GetErrorList() function.
func AccumulateFieldErrors() DecodeOption {
	return func(dec *Decoder) {
		dec.accumulateFieldErrors = true
	}
}

// Validate performs validation step in Decode()
func Validate(enabled bool) DecodeOption {
	return func(dec *Decoder) {
		dec.performValidation = enabled
	}
}

// DecodeImmutablesOnly option enforces Decoder to decode only annotations, labels and custom-encoded fields
// marked as immutable.
func DecodeImmutablesOnly() DecodeOption {
	return func(dec *Decoder) {
		dec.skipDefaultWorkload = false
		dec.filter = func(fi *fieldInfo) bool { return fi.tag.immutable || fi.tag.setOnce }
	}
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

	fun = method(out, "UnmarshalText")
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
	if isOption(out) {
		return decodeOption(out, in)
	}
	return decodePrimitive(out, in)
}

func decodeOption(out reflect.Value, in string) error {
	err := decodeUndefined(asWritableValue(out.FieldByName("value")), in)
	if err == nil {
		asWritableValue(out.FieldByName("isSet")).SetBool(true)
	}
	return err
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

	fun = method(out, "UnmarshalFromMetadata")
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

func match(values map[string]string, tag *parsedTag) string {
	if v, ok := values[tag.value]; ok {
		return v
	}
	for _, alias := range tag.aliases {
		if v, ok := values[alias]; ok {
			return v
		}
	}
	return ""
}

func (dec *Decoder) decodeField(meta *metav1.ObjectMeta, tag *parsedTag, v reflect.Value) error {
	var err error

	if tag == nil || tag.dir == out {
		return nil
	}

	if tag.inline {
		// will be handled by other structField
		return nil
	}

	switch tag.source {
	case name:
		err = decodePrimitive(v, meta.Name)
	case namespace:
		err = decodePrimitive(v, meta.Namespace)
	case label:
		err = decode(v, match(meta.Labels, tag), tag.enc)
	case annotation:
		err = decode(v, match(meta.Annotations, tag), tag.enc)
	case source(undefined):
		err = decodeCustom(v, meta)
	}

	if dec.accumulateFieldErrors && err != nil {
		dec.fieldErrors = append(dec.fieldErrors, field.TypeInvalid(field.NewPath("metadata").Child(tag.source.String()),
			tag.value, err.Error()))
	}

	if err != nil {
		return fmt.Errorf("%s '%s': [%w]", tag.source, tag.value, err)
	}

	return nil
}

func fieldByIndexWithAlloc(v reflect.Value, index []int) reflect.Value {
	if len(index) == 1 {
		return v.Field(index[0])
	}
	for i, x := range index {
		if i > 0 {
			if v.Type().Kind() == reflect.Pointer && v.Type().Elem().Kind() == reflect.Struct {
				if v.IsNil() {
					v.Set(reflect.New(v.Type().Elem()))
				}
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}
	return v
}

func (dec *Decoder) decode() error {
	return dec.iterate(func(info *fieldInfo) error {
		if !dec.filter.Apply(info) {
			return nil
		}
		if err := dec.decodeField(dec.meta, &info.tag, fieldByIndexWithAlloc(dec.root, info.path)); err != nil && !dec.accumulateFieldErrors {
			return err
		}
		return nil
	})
}

func (dec *Decoder) validate() error {
	return dec.iterate(func(info *fieldInfo) error {
		if !dec.filter.Apply(info) {
			return nil
		}
		if err := dec.validateField(dec.meta, &info.tag, fieldByIndexWithAlloc(dec.root, info.path)); err != nil && !dec.accumulateFieldErrors {
			return err
		}
		return nil
	})
}

func (dec *Decoder) iterate(fn func(info *fieldInfo) error) error {
	for _, info := range dec.cache.NameFastAccess {
		if err := fn(&info); err != nil {
			return err
		}
	}
	for _, info := range dec.cache.NamespaceFastAccess {
		if err := fn(&info); err != nil {
			return err
		}
	}
	for k := range dec.meta.Annotations {
		for _, info := range dec.cache.AnnotationFastAccess[k] {
			if err := fn(&info); err != nil {
				return err
			}
		}
	}
	for k := range dec.meta.Labels {
		for _, info := range dec.cache.LabelsFastAccess[k] {
			if err := fn(&info); err != nil {
				return err
			}
		}
	}
	for _, info := range dec.cache.CustomFieldsFastAccess {
		if err := fn(&info); err != nil {
			return err
		}
	}
	if len(dec.fieldErrors) > 0 {
		return &decodeError{message: "multiple fields errors encountered", fieldErrors: dec.fieldErrors}
	}
	return nil
}

// Decode reads data from K8s object metadata and stores them in v.
//
// See package documentation for details about deserialization.
func (dec *Decoder) Decode(meta *metav1.ObjectMeta, v any, options ...DecodeOption) error {
	// default values
	dec.fieldErrors = field.ErrorList{}
	dec.performValidation = false
	dec.accumulateFieldErrors = false
	dec.skipDefaultWorkload = false
	dec.meta = meta
	dec.filter = nil

	root := reflect.ValueOf(v)

	if root.Kind() != reflect.Pointer {
		return fmt.Errorf("required pointer to value")
	}

	if err := dec.cache.build(root); err != nil {
		return err
	}

	dec.root = dereference(root)

	for _, opt := range options {
		opt(dec)
	}

	if dec.performValidation {
		if err := dec.validate(); err != nil {
			return fmt.Errorf("failed to validate fields: %w", err)
		}
	}

	if !dec.skipDefaultWorkload {
		if err := dec.decode(); err != nil {
			return fmt.Errorf("failed to decode fields: %w", err)
		}
	}

	return nil
}

func (dec *Decoder) validateField(meta *metav1.ObjectMeta, tag *parsedTag, v reflect.Value) error {
	var err error

	// in case when setonce is used, we first check if refence values is zero. When yes
	// it is not required to validate equality
	if tag.setOnce && v.IsZero() {
		return nil
	}

	// perform equality check
	if tag.setOnce || tag.immutable {
		cv := reflect.New(v.Type()).Elem()
		if err = dec.decodeField(meta, tag, cv); err != nil {
			return fmt.Errorf("unable to decode value: [%w]", err)
		}
		if !dec.equal(v, cv) {
			err = errors.Join(err, fmt.Errorf("field is immutable"))
		}
		if dec.accumulateFieldErrors && err != nil {
			dec.fieldErrors = append(dec.fieldErrors, field.TypeInvalid(field.NewPath("metadata").Child(tag.source.String()),
				tag.value, err.Error()))
		}
		if err != nil {
			return fmt.Errorf("source: %s, key: '%s': [%w]", tag.source, tag.value, err)
		}
	}
	return nil
}

func (val *Decoder) equal(v1, v2 reflect.Value) bool {
	return reflect.DeepEqual(v1.Interface(), v2.Interface())
}

// NewDecoder returns new Decoder that reads data from meta.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Unmarshal reads data from K8s object metadata using default Decoder.
func Unmarshal(meta *metav1.ObjectMeta, v any, options ...DecodeOption) error {
	return NewDecoder().Decode(meta, v, options...)
}
