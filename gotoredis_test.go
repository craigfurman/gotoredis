package gotoredis_test

import (
	"fmt"

	"code.google.com/p/go-uuid/uuid"

	"github.com/craigfurman/gotoredis"

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

	var g *gotoredis.StructMapper

	Context("when Redis is running on supplied host and port", func() {

		BeforeEach(func() {
			var err error
			g, err = gotoredis.New("localhost:6379")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := g.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when a struct is saved", func() {

			var (
				key         string
				savedStruct SimpleStruct
				saveErr     error
			)

			BeforeEach(func() {
				key = uuid.New()

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
			})

			JustBeforeEach(func() {
				saveErr = g.Save(key, savedStruct)
			})

			It("does not error", func() {
				Expect(saveErr).ToNot(HaveOccurred())
			})

			Describe("retrieving structs", func() {

				var (
					keyToLoad       string
					retrievedStruct SimpleStruct
					retrieveErr     error
				)

				BeforeEach(func() {
					keyToLoad = key
				})

				JustBeforeEach(func() {
					retrieveErr = g.Load(keyToLoad, &retrievedStruct)
				})

				It("populates struct fields", func() {
					Expect(retrievedStruct).To(Equal(savedStruct))
				})

				It("does not error", func() {
					Expect(retrieveErr).ToNot(HaveOccurred())
				})

				Context("when there is no redis entry for key specified", func() {

					BeforeEach(func() {
						keyToLoad = "some_key_that_does_not_exist"
					})

					It("returns an error", func() {
						Expect(retrieveErr).To(MatchError(fmt.Sprintf("No Redis hash found for key %s", keyToLoad)))
					})
				})
			})

			Describe("updating structs", func() {

				var (
					updateErr error
				)

				BeforeEach(func() {
					savedStruct.String = "a new value"
				})

				JustBeforeEach(func() {
					updateErr = g.Save(key, savedStruct)
				})

				It("does not error", func() {
					Expect(updateErr).ToNot(HaveOccurred())
				})

				It("updates the struct", func() {
					var retrievedStruct SimpleStruct
					err := g.Load(key, &retrievedStruct)
					Expect(err).ToNot(HaveOccurred())
					Expect(retrievedStruct).To(Equal(savedStruct))
				})
			})

			Describe("deleting structs", func() {

				var deleteErr error

				JustBeforeEach(func() {
					deleteErr = g.Delete(key)
				})

				It("does not error", func() {
					Expect(deleteErr).NotTo(HaveOccurred())
				})

				It("cannot be retrieved from Redis", func() {
					var retrievedStruct SimpleStruct
					err := g.Load(key, &retrievedStruct)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Context("when Redis is not running on the supplied host and port", func() {

		It("returns an error", func() {
			_, err := gotoredis.New("localhost:9999")
			Expect(err).To(HaveOccurred())
		})
	})
})
