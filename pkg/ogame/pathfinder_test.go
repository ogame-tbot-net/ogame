package ogame

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathfinderSpeed(t *testing.T) {
	pf := newPathfinder()
	lf := newLfBonuses()
	assert.Equal(t, int64(12000), pf.GetSpeed(Researches{}, NoClass, lf))
	assert.Equal(t, int64(26400), pf.GetSpeed(Researches{HyperspaceDrive: 4}, NoClass, lf))
}
