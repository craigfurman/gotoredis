package gotoredis_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGotoredis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gotoredis Suite")
}
