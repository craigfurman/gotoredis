package gotoredis_test

import (
	"github.com/craigfurman/gotoredis"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SimpleStruct struct {
	String string
	Uint64 uint64
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

		Context("when a struct is saved", func() {

			var id string
			var saveErr error
			var savedStruct SimpleStruct

			BeforeEach(func() {
				savedStruct = SimpleStruct{
					String: "some string",
					Uint64: 25,
				}
				id, saveErr = g.Save(savedStruct)
			})

			It("does not error", func() {
				Expect(saveErr).ToNot(HaveOccurred())
			})

			Context("when the struct is retrieved", func() {

				var retrievedStruct SimpleStruct
				var retrieveErr error

				BeforeEach(func() {
					retrieveErr = g.Load(id, &retrievedStruct)
				})

				It("does not error", func() {
					Expect(retrieveErr).ToNot(HaveOccurred())
				})

				It("populates struct fields", func() {
					Expect(retrievedStruct).To(Equal(savedStruct))
				})
			})
		})
	})
})
