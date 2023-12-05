package watcher

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/fsnotify.v1"
)

func TestQuitOnChange(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.Nil(t, err, "failed to create temporary file")
	defer file.Close()
	defer func() {
		os.Remove(file.Name())
	}()

	var changed bool
	changedFunc := func(_ fsnotify.Event) {
		changed = true
	}

	err = quitOnChange(context.Background(), changedFunc, file.Name())
	assert.Nil(t, err, "failed to set up watcher on temporary file in %s", file)
	assert.False(t, changed, "nothing should have changed yet")

	err = os.Remove(file.Name())
	assert.Nil(t, err, "failed to delete temporary file in %s", file)

	time.Sleep(1 * time.Second)
	assert.True(t, changed, "change should have been detected")
}
