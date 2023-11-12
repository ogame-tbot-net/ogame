package ogame

type LfBonuses struct {
	Production struct {
		Metal      float64
		Crystal    float64
		Deuterium  float64
		Energy     float64
		Food       float64
		Population float64
	}
	Expeditions struct {
		Ships       float64
		Resources   float64
		Speed       float64
		DarkMatter  float64
		FleetLoss   float64
		Slots       float64
		LessEnemies float64
	}
	Dens struct {
		Metal     float64
		Crystal   float64
		Deuterium float64
	}
	Moons struct {
		Fields float64
		Size   float64
		Chance float64
	}
	Crawlers struct {
		EnergyReduction float64
		Production      float64
		Number          float64
	}
	Ships         map[ID]ShipLfBonus
	Defenses      map[ID]ShipLfBonus
	Buildings     map[ID]BaseLfBonus
	Researches    map[ID]BaseLfBonus
	LfBuildings   map[ID]BaseLfBonus
	LfResearches  map[ID]BaseLfBonus
	PhalanxRange  float64
	RecallRefund  float64
	FleetSlots    float64
	Explorations  float64
	SpaceDock     float64
	PlanetSize    float64
	InactivesLoot float64
}

func (b LfBonuses) ByShipID(id ID) ShipLfBonus {
	var tmp ShipLfBonus
	_, e := b.Ships[id]
	if e {
		tmp = b.Ships[id]
	}
	return tmp
}

func (b LfBonuses) ByDefenseID(id ID) ShipLfBonus {
	var tmp ShipLfBonus
	_, e := b.Defenses[id]
	if e {
		tmp = b.Defenses[id]
	}
	return tmp
}

func (b LfBonuses) ByUnitID(id ID) ShipLfBonus {
	if id.IsDefense() {
		return b.ByDefenseID(id)
	}
	return b.ByShipID(id)
}

func (b LfBonuses) ByBuildingID(id ID) BaseLfBonus {
	var tmp BaseLfBonus
	_, e := b.Buildings[id]
	if e {
		tmp = b.Buildings[id]
	}
	return tmp
}

func (b LfBonuses) ByTechID(id ID) BaseLfBonus {
	var tmp BaseLfBonus
	_, e := b.Researches[id]
	if e {
		tmp = b.Researches[id]
	}
	return tmp
}

func (b LfBonuses) ByLfBuildingID(id ID) BaseLfBonus {
	var tmp BaseLfBonus
	_, e := b.LfBuildings[id]
	if e {
		tmp = b.LfBuildings[id]
	}
	return tmp
}

func (b LfBonuses) ByLfTechID(id ID) BaseLfBonus {
	var tmp BaseLfBonus
	_, e := b.LfResearches[id]
	if e {
		tmp = b.LfResearches[id]
	}
	return tmp
}

func newLfBonuses() LfBonuses {
	var b LfBonuses
	b.Ships = make(map[ID]ShipLfBonus)
	b.Defenses = make(map[ID]ShipLfBonus)
	b.Buildings = make(map[ID]BaseLfBonus)
	b.Researches = make(map[ID]BaseLfBonus)
	b.LfBuildings = make(map[ID]BaseLfBonus)
	b.LfResearches = make(map[ID]BaseLfBonus)
	return b
}

type BaseLfBonus struct {
	Cost     float64
	Duration float64
}

type ShipLfBonus struct {
	Armour      float64
	Shield      float64
	Weapon      float64
	Cargo       float64
	Speed       float64
	Consumption float64
	Duration    float64
}
