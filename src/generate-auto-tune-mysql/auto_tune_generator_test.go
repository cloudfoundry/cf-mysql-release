package main_test

import (
	. "generate-auto-tune-mysql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"bytes"
)

var includeFileAt42 = `
[mysqld]
innodb_buffer_pool_size = 84
`

var includeFileAt7 = `
[mysqld]
innodb_buffer_pool_size = 6
`

var _ = Describe("AutoTuneGenerator", func() {
	Describe("Generate", func() {
		It("writes file with correct innodb buffer size", func() {
			totalMem := uint64(200)
			targetPercentage := float64(42)
			writer := &bytes.Buffer{}

			Generate(totalMem, targetPercentage, writer)

			Expect(writer.String()).To(Equal(includeFileAt42))
		})

		It("floors floating numbers to whole integers of bytes", func() {
			totalMem := uint64(10)
			targetPercentage := float64(66)
			writer := &bytes.Buffer{}

			Generate(totalMem, targetPercentage, writer)

			Expect(writer.String()).To(Equal(includeFileAt7))
		})
	})
})
