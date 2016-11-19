package metronome_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoMetronome(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoMetronome Suite")
}
