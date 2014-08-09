package gotoredis_test

import (
	"github.com/craigfurman/gotoredis"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SimpleStruct struct {
	Name string
}

var _ = Describe("saving objects in Redis", func() {

	var g *gotoredis.StructMapper

	Context("when Redis is running on expected host and port", func() {

		BeforeEach(func() {
			client, err := gotoredis.New("localhost:4567")
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

			BeforeEach(func() {
				stringValue = "some string"
				toBeSaved := SimpleStruct{
					Name: stringValue,
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

				It("rebuilds fields on struct", func() {
					Expect(retrievedObj.Name).To(Equal(stringValue))
				})
			})
		})
	})
})
