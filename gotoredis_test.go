package gotoredis_test

import (
	. "github.com/craigfurman/gotoredis"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SimpleStruct struct {
	String     string
	Int64      int64
	Int32      int32
	Int16      int16
	Int8       int8
	Int        int
	Uint64     uint64
	Uint32     uint32
	Uint16     uint16
	Uint8      uint8
	Uint       uint
	Uintptr    uintptr
	Byte       byte
	Rune       rune
	Float32    float32
	Float64    float64
	Complex64  complex64
	Complex128 complex128
	Bool       bool
}

var _ = Describe("saving objects in Redis", func() {

	var g *StructMapper

	Context("when Redis is running on supplied host and port", func() {

		BeforeEach(func() {
			var err error
			g, err = New("localhost:6379")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := g.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when a struct has already been saved in Redis", func() {

			var id string
			var savedStruct SimpleStruct

			BeforeEach(func() {
				var c64r float32 = 1.0
				var c64i float32 = 1.1
				var c128r float64 = 1.2
				var c128i float64 = 1.3
				savedStruct = SimpleStruct{
					String:     "some string",
					Int64:      64,
					Int32:      32,
					Int16:      16,
					Int8:       8,
					Int:        1000,
					Uint64:     25,
					Uint32:     9,
					Uint16:     15,
					Uint8:      10,
					Uint:       1,
					Uintptr:    77,
					Byte:       100,
					Rune:       101,
					Float32:    1.1,
					Float64:    1.2,
					Complex64:  complex(c64r, c64i),
					Complex128: complex(c128r, c128i),
					Bool:       true,
				}
				var err error
				id, err = g.Save(savedStruct)
				Expect(err).ToNot(HaveOccurred())
			})

			Describe("retrieving structs", func() {

				var retrievedStruct SimpleStruct

				BeforeEach(func() {
					err := g.Load(id, &retrievedStruct)
					Expect(err).ToNot(HaveOccurred())
				})

				It("populates struct fields", func() {
					Expect(retrievedStruct).To(Equal(savedStruct))
				})
			})

			Describe("updating structs", func() {

				BeforeEach(func() {
					savedStruct.String = "a new value"
					err := g.Update(id, savedStruct)
					Expect(err).ToNot(HaveOccurred())
				})

				It("updates the struct", func() {
					var retrievedStruct SimpleStruct
					err := g.Load(id, &retrievedStruct)
					Expect(err).ToNot(HaveOccurred())
					Expect(retrievedStruct).To(Equal(savedStruct))
				})
			})

			Describe("deleting structs", func() {

				BeforeEach(func() {
					err := g.Delete(id)
					Expect(err).ToNot(HaveOccurred())
				})

				It("cannot be retrieved from Redis", func() {
					var retrievedStruct SimpleStruct
					g.Load(id, &retrievedStruct)
					Expect(retrievedStruct).ToNot(Equal(savedStruct))
				})
			})
		})

		Describe("trying to retrieve a struct that has not been saved in Redis", func() {

			It("returns an error", func() {
				err := g.Load("foo", new(SimpleStruct))
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("when Redis is not running on the supplied host and port", func() {

		It("returns an error", func() {
			_, err := New("localhost:9999")
			Expect(err).To(HaveOccurred())
		})
	})
})
