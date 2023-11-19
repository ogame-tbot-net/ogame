package ogame

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEspionageProbeFuelConsumption(t *testing.T) {
	ep := newEspionageProbe()
	lf := newLfBonuses()
	assert.Equal(t, int64(1), ep.GetFuelConsumption(Researches{}, 1, Collector, lf))
	assert.Equal(t, int64(1), ep.GetFuelConsumption(Researches{}, 1, General, lf))
	assert.Equal(t, int64(1), ep.GetFuelConsumption(Researches{}, 0.5, Discoverer, lf))
	lf.Ships[EspionageProbeID] = ShipLfBonus{Consumption: 30}
	assert.Equal(t, int64(1), ep.GetFuelConsumption(Researches{}, 0.5, General, lf))
}

func TestEspionageProbeCargo(t *testing.T) {
	ep := newEspionageProbe()
	lf := newLfBonuses()
	assert.Equal(t, int64(0), ep.GetCargoCapacity(Researches{HyperspaceTechnology: 16}, false, NoClass, 5, lf))
	assert.Equal(t, int64(5), ep.GetCargoCapacity(Researches{HyperspaceTechnology: 2}, true, NoClass, 5, lf))
	assert.Equal(t, int64(6), ep.GetCargoCapacity(Researches{HyperspaceTechnology: 4}, true, NoClass, 5, lf))
	assert.Equal(t, int64(6), ep.GetCargoCapacity(Researches{HyperspaceTechnology: 7}, true, NoClass, 5, lf))
	assert.Equal(t, int64(9), ep.GetCargoCapacity(Researches{HyperspaceTechnology: 16}, true, NoClass, 5, lf))
}

func TestEspionageProbeSpeed(t *testing.T) {
	ep := newEspionageProbe()
	lf := newLfBonuses()
	assert.Equal(t, int64(300000000), ep.GetSpeed(Researches{CombustionDrive: 20}, General, lf))
}
