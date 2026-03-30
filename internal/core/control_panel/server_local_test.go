package controlpanel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateWaitTime(t *testing.T) {
	controlPanel := &ControlPanel{}
	waitTime := controlPanel.calculateWaitTime(0)
	assert.Equal(t, 0*time.Second, waitTime)

	waitTime = controlPanel.calculateWaitTime(3)
	assert.Equal(t, 30*time.Second, waitTime)

	waitTime = controlPanel.calculateWaitTime(8)
	assert.Equal(t, 60*time.Second, waitTime)

	waitTime = controlPanel.calculateWaitTime(15)
	assert.Equal(t, 240*time.Second, waitTime)

	waitTime = controlPanel.calculateWaitTime(16)
	assert.Equal(t, 240*time.Second, waitTime)
}
