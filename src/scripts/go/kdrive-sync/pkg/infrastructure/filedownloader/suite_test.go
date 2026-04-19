package filedownloader_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL
)

func TestFileDownloader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileDownloader Suite")
}
