package gotoredis_test

import (
	"github.com/craigfurman/gotoredis"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SimpleStruct struct {
	Name string
	Id   int
}

var _ = Describe("saving objects in Redis", func() {

	var g *gotoredis.StructMapper

	Context("when Redis is running on expected host and port", func() {

		BeforeEach(func() {
			g = gotoredis.New("localhost:4567")
		})

		Context("when a struct is saved", func() {

			var id int
			var saveErr error

			BeforeEach(func() {
				value := "some string"
				toBeSaved := SimpleStruct{
					Name: value,
				}
				id, saveErr = g.Save(toBeSaved)
			})

			It("does not error", func() {
				Expect(saveErr).ToNot(HaveOccurred())
			})

			Context("when the struct is retrieved", func() {

				var retrievedObj interface{}
				var retrieveErr error

				BeforeEach(func() {
					retrievedObj, retrieveErr = g.Load(id)
				})

				It("does not error", func() {
					Expect(retrieveErr).ToNot(HaveOccurred())
				})
			})
		})
	})
})
