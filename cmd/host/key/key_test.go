package key

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	tmpAuthFile, err := os.CreateTemp("", "authorized_keys")
	assert.Nil(t, err)

	fn := tmpAuthFile.Name()
	tmpAuthFile.Close()

	k1, err := New(fn)
	assert.Nil(t, err)

	k2, err := New(fn)
	assert.Nil(t, err)

	k3, err := New(fn)
	assert.Nil(t, err)

	assert.Nil(t, k2.Invalidate())
	assert.Nil(t, k2.Invalidate())
	assert.Nil(t, k3.Invalidate())

	contents, err := os.ReadFile(fn)
	assert.Nil(t, err)
	assert.Equal(t, contents, k1.(*key).public)

	assert.Nil(t, os.Remove(fn))
}
