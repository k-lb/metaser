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
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// test decoding name + pointer + flatten namespace
// test decoding namespace + flatten namespace`
// test decoding annotation /string/int/float/array/slice/map/time x2(pointer) + flatten
// test decoding label /string/int/float/time x2 (pointer) + flatten
// test decoding json + flatten
// test decoding textunmarshaler
// test decoding - flatten + name/namespace/annotation/label is defined
type MyStruct struct {
	A int
}

type MyStruct2 struct {
	A int
}

type MyStruct3 struct {
	A int
}

type MyStruct4 struct {
	A int64
}

func (ms *MyStruct) UnmarshalText(text []byte) error {
	ms.A = len(text)
	return nil
}

func (ms *MyStruct2) UnmarshalText(text []byte) error {
	return fmt.Errorf("Test")
}

func (ms *MyStruct4) UnmarshalFromMetadata(meta *metav1.ObjectMeta) error {
	ms.A = 0
	for k, v := range meta.Annotations {
		if strings.HasPrefix(k, "a-") {
			i, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return err
			}
			ms.A += i
		}
	}
	return nil
}

var _ = Describe("Decoder", func() {
	Context("In case when name is present in metadata", func() {
		When("decoding struct have string field with reference to name", func() {
			It("should match name from metadata", func() {
				s := struct {
					Name string `k8s:"name"`
				}{}
				m := &metav1.ObjectMeta{
					Name: "test",
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.Name).To(Equal("test"))
			})
		})
		When("decoding struct have embedded string field reference with reference to name", func() {
			It("should match name from metadata", func() {
				s := struct {
					A struct {
						Name string `k8s:"name"`
					} `k8s:"inline"`
				}{}
				m := &metav1.ObjectMeta{
					Name: "test",
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.A.Name).To(Equal("test"))
			})
		})
		When("decoding struct have string pointer field with reference to name", func() {
			It("should match name from metadata", func() {
				s := struct {
					Name *string `k8s:"name"`
				}{}
				m := &metav1.ObjectMeta{
					Name: "test",
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.Name).ToNot(BeNil())
				Expect(*s.Name).To(Equal("test"))
			})
		})
		When("decoding struct have embedded string pointer field with reference to name", func() {
			It("should match name from metadata", func() {
				s := struct {
					A struct {
						Name *string `k8s:"name"`
					} `k8s:"inline"`
				}{}
				m := &metav1.ObjectMeta{
					Name: "test",
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.A.Name).ToNot(BeNil())
				Expect(*s.A.Name).To(Equal("test"))
			})
		})
	})
	Context("In case when namespace is present in metadata", func() {
		When("decoding struct have string field with reference to namespace", func() {
			It("should match namespace from metadata", func() {
				s := struct {
					Namespace string `k8s:"namespace"`
				}{}
				m := &metav1.ObjectMeta{
					Namespace: "ns-test",
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.Namespace).To(Equal("ns-test"))
			})
		})
		When("decoding struct have embedded string field reference with reference to namespace", func() {
			It("should match namespace from metadata", func() {
				s := struct {
					A struct {
						Namespace string `k8s:"namespace"`
					} `k8s:"inline"`
				}{}
				m := &metav1.ObjectMeta{
					Namespace: "ns-test",
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.A.Namespace).To(Equal("ns-test"))
			})
		})
		When("decoding struct have string pointer field with reference to namespace", func() {
			It("should match namespace from metadata", func() {
				s := struct {
					Namespace *string `k8s:"namespace"`
				}{}
				m := &metav1.ObjectMeta{
					Namespace: "ns-test",
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.Namespace).ToNot(BeNil())
				Expect(*s.Namespace).To(Equal("ns-test"))
			})
		})
		When("decoding struct have embedded string pointer field with reference to namespace", func() {
			It("should match namespace from metadata", func() {
				s := struct {
					A struct {
						Namespace *string `k8s:"namespace"`
					} `k8s:"inline"`
				}{}
				m := &metav1.ObjectMeta{
					Namespace: "ns-test",
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.A.Namespace).ToNot(BeNil())
				Expect(*s.A.Namespace).To(Equal("ns-test"))
			})
		})
	})
	Context("In case when labels are present in metadata", func() {
		When("struct have string field with reference to label", func() {
			It("should match label from metadata", func() {
				s := struct {
					MyKey string `k8s:"label:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Labels: map[string]string{
						"mykey": "myvalue",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey).To(Equal("myvalue"))
			})
		})
	})
	Context("In case when annotations are present in metadata", func() {
		When("struct have string field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey string `k8s:"annotation:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "myvalue",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey).To(Equal("myvalue"))
			})
		})
		When("struct have string pointer field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					A struct {
						MyKey *string `k8s:"annotation:mykey"`
					} `k8s:"inline"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "myvalue",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.A.MyKey).ToNot(BeNil())
				Expect(*s.A.MyKey).To(Equal("myvalue"))
			})
		})
		When("struct have struct pointer field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					A *struct {
						MyKey string `k8s:"annotation:mykey"`
					} `k8s:"inline"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "myvalue",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.A.MyKey).ToNot(BeNil())
				Expect(s.A.MyKey).To(Equal("myvalue"))
			})
		})
		When("struct have bool field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					A struct {
						MyKey  bool `k8s:"annotation:mykey"`
						MyKey1 bool `k8s:"annotation:mykey1"`
						MyKey2 bool `k8s:"annotation:mykey2"`
						MyKey3 bool `k8s:"annotation:mykey3"`
						MyKey4 bool `k8s:"annotation:mykey4"`
						MyKey5 bool `k8s:"annotation:mykey5"`
					} `k8s:"inline"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey":  "true",
						"mykey1": "True",
						"mykey2": "1",
						"mykey3": "false",
						"mykey4": "False",
						"mykey5": "0",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.A.MyKey).To(BeTrue())
				Expect(s.A.MyKey1).To(BeTrue())
				Expect(s.A.MyKey2).To(BeTrue())
				Expect(s.A.MyKey3).To(BeFalse())
				Expect(s.A.MyKey4).To(BeFalse())
				Expect(s.A.MyKey5).To(BeFalse())
			})
		})
		When("struct have int field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey  int   `k8s:"annotation:mykey"`
					MyKey1 int8  `k8s:"annotation:mykey1"`
					MyKey2 int16 `k8s:"annotation:mykey2"`
					MyKey3 int32 `k8s:"annotation:mykey3"`
					MyKey4 int64 `k8s:"annotation:mykey4"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey":  "12",
						"mykey1": "12",
						"mykey2": "12",
						"mykey3": "12",
						"mykey4": "12",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey).To(Equal(12))
				Expect(s.MyKey1).To(Equal(int8(12)))
				Expect(s.MyKey2).To(Equal(int16(12)))
				Expect(s.MyKey3).To(Equal(int32(12)))
				Expect(s.MyKey4).To(Equal(int64(12)))
			})
		})
		When("struct have uint field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey  uint   `k8s:"annotation:mykey"`
					MyKey1 uint8  `k8s:"annotation:mykey1"`
					MyKey2 uint16 `k8s:"annotation:mykey2"`
					MyKey3 uint32 `k8s:"annotation:mykey3"`
					MyKey4 uint64 `k8s:"annotation:mykey4"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey":  "12",
						"mykey1": "12",
						"mykey2": "12",
						"mykey3": "12",
						"mykey4": "12",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey).To(Equal(uint(12)))
				Expect(s.MyKey1).To(Equal(uint8(12)))
				Expect(s.MyKey2).To(Equal(uint16(12)))
				Expect(s.MyKey3).To(Equal(uint32(12)))
				Expect(s.MyKey4).To(Equal(uint64(12)))
			})
		})
		When("struct have int pointer field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					A struct {
						MyKey *int `k8s:"annotation:mykey"`
					} `k8s:"inline"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "12",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.A.MyKey).ToNot(BeNil())
				Expect(*s.A.MyKey).To(Equal(12))
			})
		})
		When("struct have float field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey float64 `k8s:"annotation:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "12.00",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey).To(Equal(12.00))
			})
		})
		When("struct have float pointer field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					A struct {
						MyKey *float32 `k8s:"annotation:mykey"`
					} `k8s:"inline"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "12.00",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.A.MyKey).ToNot(BeNil())
				Expect(*s.A.MyKey).To(Equal(float32(12.00)))
			})
		})
		When("struct have array field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey [3]int `k8s:"annotation:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "12,1,9",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey[0]).To(Equal(12))
				Expect(s.MyKey[1]).To(Equal(1))
				Expect(s.MyKey[2]).To(Equal(9))
			})
		})
		When("struct have slice field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey []int `k8s:"annotation:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "12,1,9",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey).To(HaveLen(3))
				Expect(s.MyKey[0]).To(Equal(12))
				Expect(s.MyKey[1]).To(Equal(1))
				Expect(s.MyKey[2]).To(Equal(9))
			})
		})
		When("struct have slice field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey []int `k8s:"annotation:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).To(HaveOccurred())
			})
		})
		When("struct have map field with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey map[string]int `k8s:"annotation:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "a:12,b:1,c:9",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey).To(HaveLen(3))
				Expect(s.MyKey["a"]).To(Equal(12))
				Expect(s.MyKey["b"]).To(Equal(1))
				Expect(s.MyKey["c"]).To(Equal(9))
			})
		})
		When("struct have struct field supporting TexUnmarshler with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey MyStruct `k8s:"annotation:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "||||",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey.A).To(Equal(4))
			})
		})
		When("struct have struct field supporting TexUnmarshler with reference to annotation", func() {
			It("should return error if TextUnmarshaler returns error", func() {
				s := struct {
					MyKey MyStruct2 `k8s:"annotation:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "||||",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).To(HaveOccurred())
			})
		})
		When("struct have pre-allocated struct pointer field supporting TexUnmarshler with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				inner := MyStruct{}
				s := struct {
					MyKey *MyStruct `k8s:"annotation:mykey"`
				}{MyKey: &inner}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "||||",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(inner.A).To(Equal(4))
			})
		})
		When("struct have nil struct pointer field supporting TexUnmarshler with reference to annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey *MyStruct `k8s:"annotation:mykey"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": "||||",
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey.A).To(Equal(4))
			})
		})
		When("struct have struct pointer field with reference to json-encoded annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey *MyStruct3 `k8s:"annotation:mykey,enc:json"`
				}{MyKey: &MyStruct3{}}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": `{ "A": 12 }`,
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey.A).To(Equal(12))
			})
		})
		When("struct have struct field with reference to json-encoded annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey MyStruct3 `k8s:"annotation:mykey,enc:json"`
				}{MyKey: MyStruct3{}}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": `{ "A": 12 }`,
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey.A).To(Equal(12))
			})
		})
		When("struct have nil struct pointer field with reference to json-encoded annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey *MyStruct3 `k8s:"annotation:mykey,enc:json"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": `{ "A": 12 }`,
					},
				}
				err := Unmarshal(m, &s)
				Expect(s.MyKey).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey.A).To(Equal(12))
			})
		})
		When("struct have nil slice pointer field with reference to json-encoded annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey []string `k8s:"annotation:mykey,enc:json"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": `["A", "B", "C"]`,
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey).ToNot(BeNil())
				Expect(s.MyKey[0]).To(Equal("A"))
				Expect(s.MyKey[1]).To(Equal("B"))
				Expect(s.MyKey[2]).To(Equal("C"))
			})
		})
		When("struct have nil struct pointer field with reference to json-encoded annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey map[string][]string `k8s:"annotation:mykey,enc:json"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{
						"mykey": `{ "A": ["12", "13"], "B": ["14", "15"] }`,
					},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("struct have nil struct pointer field with reference to json-encoded annotation", func() {
			It("should match annotation from metadata", func() {
				s := struct {
					MyKey map[string]string `k8s:"annotation:mykey,enc:json"`
				}{}
				m := &metav1.ObjectMeta{
					Annotations: map[string]string{},
				}
				err := Unmarshal(m, &s)
				Expect(err).ToNot(HaveOccurred())
				Expect(s.MyKey).To(BeNil())
			})
		})
	})
	Context("In case struct tags are invalid", func() {
		It("should return error during Decode", func() {
			s := struct {
				MyKey []string `k8s:"annotation:mykey,enc:json2"`
			}{}
			m := &metav1.ObjectMeta{
				Annotations: map[string]string{
					"mykey": `["A", "B", "C"]`,
				},
			}
			err := Unmarshal(m, &s)
			Expect(err).To(HaveOccurred())
		})
	})
	Context("In case struct tags contains custom field", func() {
		It("should return be deserialized with CustomUnmarshaler ", func() {
			s := struct {
				MyKey MyStruct4 `k8s:"enc:custom"`
			}{}
			m := &metav1.ObjectMeta{
				Annotations: map[string]string{
					"a-one":   "1",
					"a-two":   "3",
					"a-three": "6",
				},
			}
			err := Unmarshal(m, &s)
			Expect(err).ToNot(HaveOccurred())
			Expect(s.MyKey.A).To(Equal(int64(10)))
		})
	})
})

var _ = Describe("Decoder with GenerateFieldsErrors enabled", func() {
	Context("In case struct tags contains field with annotation reference", func() {
		It("should return error when annotation cannot be parsed", func() {
			s := struct {
				MyKey  int     `k8s:"annotation:one"`
				MyKey2 float32 `k8s:"annotation:two"`
				MyKey3 bool    `k8s:"annotation:three"`
			}{}
			m := &metav1.ObjectMeta{
				Annotations: map[string]string{
					"one":   "abc",
					"two":   "none",
					"three": "6",
				},
			}
			dec := NewDecoder()
			err := dec.Decode(m, &s, AccumulateFieldErrors())
			Expect(err).To(HaveOccurred())
			Expect(GetErrorList(err)).ToNot(BeNil())
			Expect(GetErrorList(err)).To(HaveLen(3))
			Expect(err.Error()).ToNot(BeEmpty())
		})
	})
})

var _ = Describe("Decoder with ImmutabilityVerification enabled", func() {
	Context("In case struct tags contains field with immutable annotation reference", func() {
		It("should return no error when struct values are equal to values from annotation", func() {
			s := struct {
				MyKey  int     `k8s:"annotation:one,immutable"`
				MyKey2 float32 `k8s:"annotation:two,immutable"`
				MyKey3 bool    `k8s:"annotation:three,immutable"`
				MyKey4 string  `k8s:"annotation:four,immutable"`
			}{
				MyKey:  1,
				MyKey2: 2.0,
				MyKey3: true,
				MyKey4: "hello",
			}
			m := &metav1.ObjectMeta{
				Annotations: map[string]string{
					"one":   "1",
					"two":   "2.0",
					"three": "true",
					"four":  "hello",
				},
			}
			dec := NewDecoder()
			err := dec.Validate(m, &s)
			Expect(err).ToNot(HaveOccurred())
			Expect(s.MyKey).To(Equal(1))
			Expect(s.MyKey2).To(Equal(float32(2.0)))
			Expect(s.MyKey3).To(Equal(true))
			Expect(s.MyKey4).To(Equal("hello"))
		})
		It("should return errors when struct values are not equal to values from annotation", func() {
			s := struct {
				MyKey  int     `k8s:"annotation:one,immutable"`
				MyKey2 float32 `k8s:"annotation:two,immutable"`
				MyKey3 bool    `k8s:"annotation:three,immutable"`
				MyKey4 string  `k8s:"annotation:four,immutable"`
			}{
				MyKey:  2,
				MyKey2: 3.0,
				MyKey3: false,
				MyKey4: "hello2",
			}
			m := &metav1.ObjectMeta{
				Annotations: map[string]string{
					"one":   "1",
					"two":   "2.0",
					"three": "true",
					"four":  "hello",
				},
			}
			dec := NewDecoder()
			err := dec.Validate(m, &s, AccumulateFieldErrors())
			Expect(err).To(HaveOccurred())
			Expect(s.MyKey).To(Equal(2))
			Expect(s.MyKey2).To(Equal(float32(3.0)))
			Expect(s.MyKey3).To(Equal(false))
			Expect(s.MyKey4).To(Equal("hello2"))
			Expect(GetErrorList(err)).ToNot(BeNil())
			Expect(GetErrorList(err)).To(HaveLen(4))
			Expect(err.Error()).ToNot(BeEmpty())
		})
	})
})

var _ = Describe("Decoder with ImmutabilityVerification disabled", func() {
	It("should return no errors when struct values are not equal to values from annotation", func() {
		s := struct {
			MyKey  int     `k8s:"annotation:one,immutable"`
			MyKey2 float32 `k8s:"annotation:two,immutable"`
			MyKey3 bool    `k8s:"annotation:three,immutable"`
			MyKey4 string  `k8s:"annotation:four,immutable"`
		}{
			MyKey:  2,
			MyKey2: 3.0,
			MyKey3: false,
			MyKey4: "hello2",
		}
		m := &metav1.ObjectMeta{
			Annotations: map[string]string{
				"one":   "1",
				"two":   "2.0",
				"three": "true",
				"four":  "hello",
			},
		}
		dec := NewDecoder()
		err := dec.Decode(m, &s, AccumulateFieldErrors())
		Expect(err).ToNot(HaveOccurred())
		Expect(s.MyKey).To(Equal(1))
		Expect(s.MyKey2).To(Equal(float32(2.0)))
		Expect(s.MyKey3).To(Equal(true))
		Expect(s.MyKey4).To(Equal("hello"))
	})
})

var _ = Describe("Decoding embedded structs", func() {
	It("should return no errors", func() {
		type DummyA struct {
			A string
		}
		type DummyB struct {
			A string
		}
		type DummyC struct {
			A string
		}
		type A struct {
			DummyA
			Z int
			X string `k8s:"annotation:iks"`
			V int
		}
		type B struct {
			DummyB
			Z int
			A `k8s:"inline"`
			V int
		}
		type C struct {
			Z int
			B `k8s:"inline"`
			V int
		}
		type D struct {
			DummyC
			DummyA
			Z int
			C `k8s:"inline"`
			V int
		}
		m := &metav1.ObjectMeta{
			Annotations: map[string]string{
				"iks": "test",
			},
		}
		v := D{}
		dec := NewDecoder()
		err := dec.Decode(m, &v)
		Expect(err).ToNot(HaveOccurred())
		Expect(v.C.B.A.X).To(Equal("test"))
	})
})

var _ = Describe("Decoding Option struct", func() {
	type A struct {
		X Option[*bool] `k8s:"annotation:iks"`
	}
	When("annotation exists", func() {
		It("Should set proper Option values when annotation value is true", func() {
			v := A{}
			m := &metav1.ObjectMeta{
				Annotations: map[string]string{
					"iks": "true",
				},
			}
			dec := NewDecoder()
			err := dec.Decode(m, &v)
			Expect(err).ToNot(HaveOccurred())
			Expect(v.X.IsSet()).To(BeTrue())
			Expect(v.X.Get()).ToNot(BeNil())
			Expect(*v.X.Get()).To(BeTrue())
		})
		It("Should set proper Option values when annotation value is false", func() {
			v := A{}
			m := &metav1.ObjectMeta{
				Annotations: map[string]string{
					"iks": "false",
				},
			}
			dec := NewDecoder()
			err := dec.Decode(m, &v)
			Expect(err).ToNot(HaveOccurred())
			Expect(v.X.IsSet()).To(BeTrue())
			Expect(v.X.Get()).ToNot(BeNil())
			Expect(*v.X.Get()).To(BeFalse())
		})
	})
	When("annotation does not exist", func() {
		It("Should set proper Option values", func() {
			v := A{}
			m := &metav1.ObjectMeta{
				Annotations: map[string]string{},
			}
			dec := NewDecoder()
			err := dec.Decode(m, &v)
			Expect(err).ToNot(HaveOccurred())
			Expect(v.X.IsSet()).To(BeFalse())
		})
	})
})
