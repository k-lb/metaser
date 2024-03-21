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
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MyStruct5 struct {
	A []int
}

func (ms *MyStruct5) MarshalToMetadata(meta *metav1.ObjectMeta) error {
	i := 1
	for _, v := range ms.A {
		meta.Annotations[fmt.Sprintf("a-%d", i)] = strconv.FormatInt(int64(v), 10)
		i++
	}
	return nil
}

type MyStruct6 struct {
	A []int
}

func (ms *MyStruct6) MarshalText() ([]byte, error) {
	str := "vals-"
	for _, v := range ms.A {
		str += strconv.FormatInt(int64(v), 10) + ";"
	}
	return []byte(str), nil
}

var _ = Describe("Encoder", func() {
	Context("In case when name is present in metadata", func() {
		When("struct's string field have name reference", func() {
			It("should encode name in metadata", func() {
				s := struct {
					Name string `k8s:"name"`
				}{Name: "test"}
				m := &metav1.ObjectMeta{}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Name).To(Equal("test"))
			})
		})
		When("struct's field have namespace reference", func() {
			It("should encode namespace in metadata", func() {
				s := struct {
					Name string `k8s:"namespace"`
				}{Name: "test"}
				m := &metav1.ObjectMeta{}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Namespace).To(Equal("test"))
			})
		})
		When("struct's field have label reference", func() {
			It("should encode label in metadata", func() {
				s := struct {
					Name string `k8s:"label:testkey"`
				}{Name: "test"}
				m := &metav1.ObjectMeta{Labels: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Labels).To(HaveKeyWithValue("testkey", "test"))
			})
		})
		When("struct's string field have annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value string `k8s:"annotation:testkey"`
				}{Value: "test"}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKeyWithValue("testkey", "test"))
			})
		})
		When("struct's int fields have annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value  int   `k8s:"annotation:testkey"`
					Value1 int8  `k8s:"annotation:testkey1"`
					Value2 int16 `k8s:"annotation:testkey2"`
					Value3 int32 `k8s:"annotation:testkey3"`
					Value4 int64 `k8s:"annotation:testkey4"`
				}{
					Value:  -1,
					Value1: -1,
					Value2: -1,
					Value3: -1,
					Value4: -1,
				}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKeyWithValue("testkey", "-1"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey1", "-1"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey2", "-1"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey3", "-1"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey4", "-1"))
			})
		})
		When("struct's uint fields have annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value  uint   `k8s:"annotation:testkey"`
					Value1 uint8  `k8s:"annotation:testkey1"`
					Value2 uint16 `k8s:"annotation:testkey2"`
					Value3 uint32 `k8s:"annotation:testkey3"`
					Value4 uint64 `k8s:"annotation:testkey4"`
				}{
					Value:  1,
					Value1: 1,
					Value2: 1,
					Value3: 1,
					Value4: 1,
				}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKeyWithValue("testkey", "1"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey1", "1"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey2", "1"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey3", "1"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey4", "1"))
			})
		})
		When("struct's bool fields have annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value  bool `k8s:"annotation:testkey"`
					Value1 bool `k8s:"annotation:testkey1"`
				}{Value: true, Value1: false}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKeyWithValue("testkey", "true"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey1", "false"))
			})
		})
		When("struct's float fields have annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value  float32 `k8s:"annotation:testkey"`
					Value1 float64 `k8s:"annotation:testkey1"`
				}{Value: 1.66, Value1: 1.33}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKeyWithValue("testkey", "1.66"))
				Expect(m.Annotations).To(HaveKeyWithValue("testkey1", "1.33"))
			})
		})
		When("encoded struct field have input-only annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value string `k8s:"annotation:testkey,in"`
				}{Value: "test"}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).NotTo(HaveKeyWithValue("testkey", "test"))
			})
		})
		When("encoded struct field have input-only annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value []string `k8s:"annotation:testkey"`
				}{Value: []string{"test", "value", "sample"}}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKeyWithValue("testkey", "test,value,sample"))
			})
		})
		When("encoded struct field have input-only annotation reference", func() {
			It("should encode annotation in metadata", func() {
				str := "test"
				s := struct {
					Value *string `k8s:"annotation:testkey"`
				}{Value: &str}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKeyWithValue("testkey", "test"))
			})
		})
		When("encoded struct field have input-only annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value *string `k8s:"annotation:testkey"`
				}{Value: nil}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKey("testkey"))
			})
		})
		When("encoded struct field have input-only annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value *string `k8s:"annotation:testkey,omitempty"`
				}{Value: nil}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).NotTo(HaveKey("testkey"))
			})
		})
		When("encoded struct field have input-only annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value map[string]string `k8s:"annotation:testkey,omitempty"`
				}{Value: map[string]string{"A": "a", "B": "b"}}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKey("testkey"))
				Expect(m.Annotations["testkey"]).To(BeElementOf([]string{"A:a,B:b", "B:b,A:a"}))
			})
		})
		When("encoded struct field have input-only annotation reference", func() {
			It("should encode annotation in metadata", func() {
				s := struct {
					Value map[string]string `k8s:"annotation:testkey,omitempty"`
				}{}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).ToNot(HaveKey("testkey"))
			})
		})
	})
	Context("In case struct tags contains custom field", func() {
		It("should return be serialized with MetadataMarshaler ", func() {
			s := struct {
				MyKey MyStruct5 `k8s:"custom"`
			}{
				MyKey: MyStruct5{
					A: []int{1, 3, 6},
				},
			}
			m := &metav1.ObjectMeta{Annotations: map[string]string{}}
			err := Marshal(&s, m)
			Expect(err).ToNot(HaveOccurred())
			Expect(m.Annotations).To(HaveKeyWithValue("a-1", "1"))
			Expect(m.Annotations).To(HaveKeyWithValue("a-2", "3"))
			Expect(m.Annotations).To(HaveKeyWithValue("a-3", "6"))
		})
	})
	Context("In case struct contains field implementing TextMarshaler", func() {
		It("should have annotation string matching TextMarshaler", func() {
			s := struct {
				MyKey MyStruct6 `k8s:"annotation:test"`
			}{
				MyKey: MyStruct6{
					A: []int{1, 3, 6},
				},
			}
			m := &metav1.ObjectMeta{Annotations: map[string]string{}}
			err := Marshal(&s, m)
			Expect(err).ToNot(HaveOccurred())
			Expect(m.Annotations).To(HaveKeyWithValue("test", "vals-1;3;6;"))
		})
	})
	Context("In case struct tags contains json-encoded field", func() {
		It("should return be serialized as valid json", func() {
			s := struct {
				MyKey MyStruct5 `k8s:"annotation:test,enc:json"`
			}{
				MyKey: MyStruct5{
					A: []int{1, 3, 6},
				},
			}
			m := &metav1.ObjectMeta{Annotations: map[string]string{}}
			err := Marshal(&s, m)
			Expect(err).ToNot(HaveOccurred())
			Expect(m.Annotations).To(HaveKeyWithValue("test", `{"A":[1,3,6]}`))
		})
	})
	Context("In case struct contains Option field", func() {
		type s struct {
			MyKey Option[bool] `k8s:"annotation:test,omitempty"`
		}
		Context("Option is set to true", func() {
			It("should be serialized", func() {
				s := s{
					MyKey: Some(true),
				}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKeyWithValue("test", "true"))
			})
		})
		Context("Option is set to false", func() {
			It("should be serialized", func() {
				s := s{
					MyKey: Some(false),
				}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).To(HaveKeyWithValue("test", "false"))
			})
		})
		Context("Option is not set", func() {
			It("should not be serialized", func() {
				s := s{
					MyKey: None[bool](),
				}
				m := &metav1.ObjectMeta{Annotations: map[string]string{}}
				err := Marshal(&s, m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Annotations).ToNot(HaveKey("test"))
			})
		})
	})
})
