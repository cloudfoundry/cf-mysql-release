package mariadb_start_manager_test

import (

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MariadbStartManager", func() {
	Describe("When deploying a single node cluster", func(){
		Context("On first deploy", func() {
			Context("When upgrade is necessary after restart", func(){

			})
			Context("When upgrade is NOT necessary after restart", func(){

			})
		})
		Context("When last deploy had multiple nodes", func() {
			Context("When upgrade is necessary after restart", func(){

			})
			Context("When upgrade is NOT necessary after restart", func(){

			})
		})
		Context("When last deploy had a single node", func() {
			Context("When upgrade is necessary after restart", func(){

			})
			Context("When upgrade is NOT necessary after restart", func(){

			})
		})
	})
	Describe("When deploying a cluster with mutple nodes", func(){
			Context("On first deploy", func() {
					Context("When upgrade is necessary after restart", func(){

						})
					Context("When upgrade is NOT necessary after restart", func(){

						})
				})
			Context("When last deploy had multiple nodes", func() {
					Context("When upgrade is necessary after restart", func(){

						})
					Context("When upgrade is NOT necessary after restart", func(){

						})
				})
			Context("When last deploy had a single node", func() {
					Context("When upgrade is necessary after restart", func(){

						})
					Context("When upgrade is NOT necessary after restart", func(){

						})
				})
		})
})
