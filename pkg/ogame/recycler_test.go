package ogame

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecyclerSpeed(t *testing.T) {
	r := newRecycler()
	lf := newLfBonuses()
	assert.Equal(t, int64(3200), r.GetSpeed(Researches{CombustionDrive: 6, ImpulseDrive: 1, HyperspaceDrive: 1}, NoClass, lf))
	assert.Equal(t, int64(17600), r.GetSpeed(Researches{CombustionDrive: 1, ImpulseDrive: 17, HyperspaceDrive: 10}, NoClass, lf))
	assert.Equal(t, int64(33000), r.GetSpeed(Researches{CombustionDrive: 1, ImpulseDrive: 17, HyperspaceDrive: 15}, NoClass, lf))
	assert.Equal(t, int64(18400), r.GetSpeed(Researches{CombustionDrive: 1, ImpulseDrive: 18, HyperspaceDrive: 10}, NoClass, lf))
	assert.Equal(t, int64(34800), r.GetSpeed(Researches{CombustionDrive: 1, ImpulseDrive: 17, HyperspaceDrive: 16}, NoClass, lf))
	assert.Equal(t, int64(42600), r.GetSpeed(Researches{CombustionDrive: 1, ImpulseDrive: 17, HyperspaceDrive: 17}, General, lf))
}

func TestRecyclerCargo(t *testing.T) {
	r := newRecycler()
	lf := newLfBonuses()
	assert.Equal(t, int64(40000), r.GetCargoCapacity(Researches{HyperspaceTechnology: 16}, false, General, 5, lf))
	lf.Ships[RecyclerID] = ShipLfBonus{Cargo: 59.192}
	assert.Equal(t, int64(51838), r.GetCargoCapacity(Researches{HyperspaceTechnology: 16}, false, General, 5, lf))
}

func TestRecyclerConsumption(t *testing.T) {
	r := newRecycler()
	lf := newLfBonuses()
	assert.Equal(t, int64(900), r.GetFuelConsumption(Researches{HyperspaceDrive: 17}, 1, NoClass, lf))
	assert.Equal(t, int64(675), r.GetFuelConsumption(Researches{HyperspaceDrive: 17}, 1, General, lf))
}
