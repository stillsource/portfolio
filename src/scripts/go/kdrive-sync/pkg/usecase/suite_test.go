package usecase_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL is the idiomatic style
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL is the idiomatic style
)

func TestUsecase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Usecase Suite")
}
