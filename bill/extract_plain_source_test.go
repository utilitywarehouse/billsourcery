package bill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTargetFilenameExtn(t *testing.T) {
	assert := assert.New(t)

	pse := &plainSourceExtractor{
		sourceRoot: "/source/dir/",
		targetDir:  "/target/dir/",
	}

	assert.Equal("/target/dir/a/b/c_1.ext", pse.targetName("/source/dir/a/b/c.ext", 1))
}

func TestTargetFilenameNoExtn(t *testing.T) {
	assert := assert.New(t)

	pse := &plainSourceExtractor{
		sourceRoot: "/source/dir/",
		targetDir:  "/target/dir/",
	}

	assert.Equal("/target/dir/a/b/c_1", pse.targetName("/source/dir/a/b/c", 1))
}
