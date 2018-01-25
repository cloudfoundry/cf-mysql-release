package config_test

import (
	. "quota-enforcer/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Validate", func() {
		var config Config

		BeforeEach(func() {
			config = Config{
				Host:           "fake-host",
				Port:           9999,
				User:           "fake-user",
				Password:       "fake-password",
				IgnoredUsers:   []string{"fake-ignored-user"},
				DBName:         "fake-db-name",
				PauseInSeconds: 1,
			}
		})

		It("validates a valid config file", func() {
			err := config.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when Host is not specified", func() {
			BeforeEach(func() {
				config.Host = ""
			})

			It("returns a validation error", func() {
				err := config.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Host"))
			})
		})

		Context("when Port is not specified", func() {
			BeforeEach(func() {
				config.Port = 0
			})

			It("returns a validation error", func() {
				err := config.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Port"))
			})
		})

		Context("when User is not specified", func() {
			BeforeEach(func() {
				config.User = ""
			})

			It("returns a validation error", func() {
				err := config.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("User"))
			})
		})

		Context("when IgnoredUsers is not specified", func() {
			BeforeEach(func() {
				config.IgnoredUsers = []string{}
			})

			It("does not return a validation error", func() {
				err := config.Validate()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when DBName is not specified", func() {
			BeforeEach(func() {
				config.DBName = ""
			})

			It("returns a validation error", func() {
				err := config.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("DBName"))
			})
		})

		Context("when PauseInSeconds is not specified", func() {
			BeforeEach(func() {
				config.PauseInSeconds = 0
			})

			It("returns a validation error", func() {
				err := config.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("PauseInSeconds"))
			})
		})

		Context("when PauseInSeconds is negative", func() {
			BeforeEach(func() {
				config.PauseInSeconds = -1
			})

			It("returns a validation error", func() {
				err := config.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("PauseInSeconds"))
			})
		})

	})
})
