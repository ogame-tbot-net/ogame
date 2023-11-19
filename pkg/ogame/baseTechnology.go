package ogame

import (
	"math"
	"time"
)

// BaseTechnology base struct for technologies
type BaseTechnology struct {
	BaseLevelable
}

// TechnologyConstructionTime returns the duration it takes to build given technology
func (b BaseTechnology) TechnologyConstructionTime(level, universeSpeed int64, acc TechAccelerators, hasTechnocrat bool, class CharacterClass) time.Duration {
	price := b.GetPrice(level)
	metalCost := float64(price.Metal)
	crystalCost := float64(price.Crystal)
	researchLabLvl := float64(acc.GetResearchLab())
	hours := (metalCost + crystalCost) / (1000 * (1 + researchLabLvl) * float64(universeSpeed))
	if hasTechnocrat {
		hours -= 0.25 * hours
	}
	if class == Discoverer {
		hours -= 0.25 * hours
	}
	secs := math.Max(1, hours*3600)
	return time.Duration(int64(math.Floor(secs))) * time.Second
}

// ConstructionTime same as TechnologyConstructionTime, needed for BaseOgameObj implementation
func (b BaseTechnology) ConstructionTime(level, universeSpeed int64, facilities BuildAccelerators, hasTechnocrat bool, class CharacterClass) time.Duration {
	return b.TechnologyConstructionTime(level, universeSpeed, facilities, hasTechnocrat, class)
}

// ConstructionTimeWithBonuses returns duration with LfBonuses applied
func (b BaseTechnology) ConstructionTimeWithBonuses(level, universeSpeed int64, facilities BuildAccelerators, hasTechnocrat bool, class CharacterClass, bonuses LfBonuses) time.Duration {
	duration := b.TechnologyConstructionTime(level, universeSpeed, facilities, hasTechnocrat, class)
	bonus := bonuses.ByUnitID(b.ID).Duration
	return time.Duration(float64(duration) - float64(duration)*bonus)
}

// GetLevel returns current level of a technology
func (b BaseTechnology) GetLevel(_ IResourcesBuildings, _ IFacilities, researches IResearches) int64 {
	return researches.ByID(b.ID)
}

// GetPriceWithBonus return the price with LfBonuses applied
func (b BaseTechnology) GetPriceWithBonuses(level int64, bonuses LfBonuses) Resources {
	price := b.GetPrice(level)
	bonus := bonuses.ByTechID(b.ID).Cost
	return price.SubPercent(bonus)
}
