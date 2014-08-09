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
			g = gotoredis.New("localhost:4567")
		})

		Context("when a struct is saved", func() {

			var err error

			BeforeEach(func() {
				value := "some string"
				toBeSaved := SimpleStruct{
					Name: value,
				}
				err = g.Save(toBeSaved)
			})

			It("does not error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
