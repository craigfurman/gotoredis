package gotoredis_test

import (
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	redisWorkingDir string
	redisServer     *gexec.Session
)

func TestGotoredis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gotoredis Suite")
}

var _ = BeforeSuite(func() {
	var err error
	redisWorkingDir, err = ioutil.TempDir("", "redis")
	Expect(err).ToNot(HaveOccurred())
	cmd := exec.Command("redis-server", "--dir", redisWorkingDir)
	redisServer, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(func() bool {
		conn, err := net.Dial("tcp", "localhost:6379")
		if err == nil {
			conn.Close()
			return true
		}
		return false
	}).Should(BeTrue())
})

var _ = AfterSuite(func() {
	redisServer.Terminate().Wait(time.Second * 5)
	Eventually(redisServer).Should(gexec.Exit())
	err := os.RemoveAll(redisWorkingDir)
	Expect(err).ToNot(HaveOccurred())
})
