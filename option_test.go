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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Option", func() {
	Context("Value is Some", func() {
		It("", func() {
			v := Some(true)
			Expect(v.IsSet()).To(BeTrue())
			Expect(v.Get()).To(BeTrue())
			Expect(v.GetOrDefault(true)).To(BeTrue())
			Expect(v.GetOrDefault(false)).To(BeTrue())

			v = Some(false)
			Expect(v.IsSet()).To(BeTrue())
			Expect(v.Get()).To(BeFalse())
			Expect(v.GetOrDefault(true)).To(BeFalse())
			Expect(v.GetOrDefault(false)).To(BeFalse())
		})
	})
	Context("Value is None", func() {
		It("", func() {
			v := None[bool]()
			Expect(v.IsSet()).To(BeFalse())
			Expect(func() { v.Get() }).To(Panic())
			Expect(v.GetOrDefault(true)).To(BeTrue())
			Expect(v.GetOrDefault(false)).To(BeFalse())
		})
	})
})
