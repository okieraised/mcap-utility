package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListMCAPFilesInDirectory(t *testing.T) {
	files, err := ListMCAPFilesInDirectory("/home/pham/workspace/repo/mcap-utility/example")
	assert.NoError(t, err)

	for _, file := range files {
		fmt.Println(file)
	}
}
