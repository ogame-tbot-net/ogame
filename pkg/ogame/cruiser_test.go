package ogame

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCruiser_RapidfireAgainst(t *testing.T) {
	c := newCruiser()
	assert.Equal(t, map[ID]int64{EspionageProbeID: 5, SolarSatelliteID: 5, LightFighterID: 6, RocketLauncherID: 10, CrawlerID: 5}, c.GetRapidfireAgainst())
}

func TestCruiser_GetCargoCapacity(t *testing.T) {
	c := newCruiser()
	l := newLfBonuses()
	assert.Equal(t, int64(800), c.GetCargoCapacity(Researches{HyperspaceTechnology: 0}, false, NoClass, 5, l))
	assert.Equal(t, int64(1120), c.GetCargoCapacity(Researches{HyperspaceTechnology: 8}, false, NoClass, 5, l))
	l.Ships[CruiserID] = ShipLfBonus{Cargo: 22.197}
	assert.Equal(t, int64(1617), c.GetCargoCapacity(Researches{HyperspaceTechnology: 16}, false, NoClass, 5, l))
}

func TestCruiser_GetSpeed(t *testing.T) {
	c := newCruiser()
	l := newLfBonuses()
	l.Ships[CruiserID] = ShipLfBonus{Speed: 22.197}
	assert.Equal(t, int64(84330), c.GetSpeed(Researches{ImpulseDrive: 17}, General, l))
}

func TestCruiser_GetFuelConsumption(t *testing.T) {
	c := newCruiser()
	l := newLfBonuses()
	assert.Equal(t, int64(300), c.GetFuelConsumption(Researches{}, 1, Collector, l))
}

func TestCruiser_GetPrice(t *testing.T) {
	c := newCruiser()
	assert.Equal(t, Resources{Metal: 20000, Crystal: 7000, Deuterium: 2000}, c.GetPrice(1))
	assert.Equal(t, Resources{Metal: 60000, Crystal: 21000, Deuterium: 6000}, c.GetPrice(3))
}

func TestCruiser_GetShield(t *testing.T) {
	c := newCruiser()
	l := newLfBonuses()
	assert.Equal(t, int64(145), c.GetShieldPower(Researches{ShieldingTechnology: 19}, l))
	l.Ships[CruiserID] = ShipLfBonus{Shield: 22.197}
	assert.Equal(t, int64(156), c.GetShieldPower(Researches{ShieldingTechnology: 19}, l))
}
