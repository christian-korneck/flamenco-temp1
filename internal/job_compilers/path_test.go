package job_compilers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStem(t *testing.T) {
	assert.Equal(t, "", Stem(""))
	assert.Equal(t, "stem", Stem("stem.txt"))
	assert.Equal(t, "stem.a", Stem("stem.a.b"))
	assert.Equal(t, "file", Stem("/path/to/file.txt"))
	// assert.Equal(t, "file", Stem("C:\\path\\to\\file.txt"))
}
