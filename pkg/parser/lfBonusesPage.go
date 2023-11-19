package parser

import "github.com/alaingilbert/ogame/pkg/ogame"

func (p LfBonusesPage) ExtractLfBonuses() (ogame.LfBonuses, error) {
	return p.e.ExtractLfBonusesFromDoc(p.GetDoc())
}

func (p LfBonusesPage) ExtractPlanetID() (ogame.CelestialID, error) {
	return p.e.ExtractPlanetID(p.content)
}
