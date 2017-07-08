package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMitm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mitm Suite")
}
