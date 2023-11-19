package ogame

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLargeCargo_GetSpeed(t *testing.T) {
	lc := newLargeCargo()
	lf := newLfBonuses()
	assert.Equal(t, int64(12000), lc.GetSpeed(Researches{CombustionDrive: 6}, NoClass, lf))
	assert.Equal(t, int64(19500), lc.GetSpeed(Researches{CombustionDrive: 6}, Collector, lf))
}

func TestLargeCargo_GetCargoCapacity(t *testing.T) {
	lc := newLargeCargo()
	lf := newLfBonuses()
	assert.Equal(t, int64(35000), lc.GetCargoCapacity(Researches{HyperspaceTechnology: 8}, false, NoClass, 5, lf))
	assert.Equal(t, int64(37500), lc.GetCargoCapacity(Researches{HyperspaceTechnology: 10}, false, NoClass, 5, lf))
	assert.Equal(t, int64(43750), lc.GetCargoCapacity(Researches{HyperspaceTechnology: 10}, false, Collector, 5, lf))
	lf.Ships[LargeCargoID] = ShipLfBonus{Cargo: 59.192}
	assert.Equal(t, int64(59798), lc.GetCargoCapacity(Researches{HyperspaceTechnology: 16}, false, General, 5, lf))
}
