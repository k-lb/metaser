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

// package metaser implements serialization and deserialization of structs into/from Kubernetes object's metadata.
//
// The primary use-case of this library is to allow easy storing of user-defined data structures in Kubernetes object's annotations and labels.
//
// Currently only 'name', 'namespace', 'labels', 'annotations' fields from Kubernets object's metadata are supported by metaser.
// metaser is using 'k8s' structure tag prefix for definiting serialization scheme.
//
// Example:
//
//	type MyData struct {
//	  MyAnnotationVal int     `k8s:"annotation:myannotation,omitempty"`
//	  MyLabelVal      float32 `k8s:"label:mylabel,in"`
//	  MyNameVal       string  `k8s:"name,in"`
//	}
//
// Supported tag values:
//   - annotation - indicate if field should be serialized/deserialized from k8s Annotations map. The annotation should follow "annotation:<key>" syntax, where <key> should be valid k8s [annotation]
//   - label - indicate if field should be serialized/deserialized from k8s Labels map. The annotation should follow "label:<key>" syntax, where <key> should be valid k8s [label]
//   - name - indicate if field should be serialized/deserialized from k8s Name value.
//   - namespace - indicate if field should be serialized/deserialized from k8s Namespace field value.
//   - enc - sets encoding/decoding scheme for field. If ommited default schema will be used (see Supported types section for more info). If type is not in supported type list the TextMarshaler/TextUnmarshaler will be used. Tag should follow enc:<val> syntax, where val is one of supported values defined in Encoding schemes section.
//   - in - indicate if field should be used during decoding and ignored during encoding
//   - inout - indicate if field should be used during decoding and encoding. This is default value if 'in' or 'out' is not set explicitly.
//   - out - indicate if field should be used during encoding and ignored during decoding
//   - inline - can be only used on struct fields. Inline all contained structure fields into outer struct.
//   - omitempty - do not encode field if have zero value. If the annotation or label exists it will be removed from metadata.
//   - immutable - the value of field cannot change during decoding.
//   - aliases - specify alternative keys for annotations or labels loopkup. Can be used only with 'annotation' or 'label' tag. The tag have following syntax: 'aliases:value1;value2;value3'. Values should be a valid k8s annotation or label key.
//
// Encoding schemes:
//   - json - field will deserialized/serialized with json decoder/encoder
//   - custom - field will be deserialized/serialized with metaser.MetadataUnmarshaler/metaser.MetadataMarshaler interface.
//
// Supported types:
//   - bool - serialized/deserialized using strconv package.
//   - int, int8, int16, int32, int64 - serialized/deserialized using strconv package.
//   - uint, uint8, uint16, uint32, uint64 - serialized/deserialized using strconv package.
//   - float32, float64  - serialized/deserialized using strconv package.
//   - string
//   - array - encodes field as comma separated list of elements. Serialized elements cannot contain comma.
//   - slice - encodes field as comma separated list of elements. Serialized elements cannot contain comma.
//   - map - encodes field as comma separated list of <key>:<value> pairs. Serialized elements cannot contain comma or semicolon.
//   - pointers - pointers will be dereferenced during serialization/deserialization.
//   - struct - structs can be only used with 'inline' tag.
//   - metaser.Option[T] - generic struct representing optional value.
//
// Limitations:
//   - current implemntation does not support reference cycles inside decoded and encoded structs. The result of such operations is undefined.
//
// [annotation]: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/#syntax-and-character-set
// [label]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
package metaser
