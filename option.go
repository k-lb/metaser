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

const (
	valueFieldIndex = 0
	isSetFieldIndex = 1
)

// Option is a type representing optional value that may or may not be present.
type Option[T any] struct {
	value T
	isSet bool
}

// Some constructs new Option with value set to 'value'.
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, isSet: true}
}

// None constructs new Option without value.
func None[T any]() Option[T] {
	return Option[T]{isSet: false}
}

// Get gets contained value. If the values was not set it panics.
func (s *Option[T]) Get() T {
	if s.isSet {
		return s.value
	}
	panic("Option value is not set.")
}

// GetOrElse gets contained value. If the values was not set it returs other.
func (s *Option[T]) GetOrDefault(def T) T {
	if s.isSet {
		return s.value
	}
	return def
}

// IsSet validates if internal option value was set.
func (s *Option[_]) IsSet() bool {
	return s.isSet
}
