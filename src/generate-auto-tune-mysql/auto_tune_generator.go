package main

import (
	"io"
	"fmt"
)

func Generate(totalMem uint64, targetPercentage float64, writer io.Writer) {

	bufferPoolSize := float64(totalMem) * targetPercentage / 100.0
	writer.Write([]byte(fmt.Sprintf(`
[mysqld]
innodb_buffer_pool_size = %d
`, uint64(bufferPoolSize))))
}
