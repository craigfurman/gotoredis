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
			var stringValue string
			var uint64Value uint64

			BeforeEach(func() {
				stringValue = "some string"
				uint64Value = 25
				toBeSaved := SimpleStruct{
					String: stringValue,
					Uint64: uint64Value,
				}
				id, saveErr = g.Save(toBeSaved)
			})

			It("does not error", func() {
				Expect(saveErr).ToNot(HaveOccurred())
			})

			Context("when the struct is retrieved", func() {

				var retrievedObj SimpleStruct
				var retrieveErr error

				BeforeEach(func() {
					retrieveErr = g.Load(id, &retrievedObj)
				})

				It("does not error", func() {
					Expect(retrieveErr).ToNot(HaveOccurred())
				})

				It("retrieves string values", func() {
					Expect(retrievedObj.String).To(Equal(stringValue))
				})

				It("retrieves uint64 values", func() {
					Expect(retrievedObj.Uint64).To(Equal(uint64Value))
				})
			})
		})
	})
})
