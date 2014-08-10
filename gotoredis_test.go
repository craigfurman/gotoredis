package gotoredis_test

import (
	"github.com/craigfurman/gotoredis"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SimpleStruct struct {
	String  string
	Uint64  uint64
	Uint32  uint32
	Uint16  uint16
	Uint8   uint8
	Uint    uint
	Uintptr uintptr
	Bool    bool
}

var _ = Describe("saving objects in Redis", func() {

	var g *gotoredis.StructMapper

	Context("when Redis is running on expected host and port", func() {

		BeforeEach(func() {
			client, err := gotoredis.New("localhost:6379")
			g = client
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
				savedStruct = SimpleStruct{
					String:  "some string",
					Uint64:  25,
					Uint32:  9,
					Uint16:  15,
					Uint8:   10,
					Uint:    1,
					Uintptr: 77,
					Bool:    true,
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
	})
})
