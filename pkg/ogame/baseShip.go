package ogame

import "math"

// BaseShip base struct for ships
type BaseShip struct {
	BaseDefender
	BaseCargoCapacity int64
	BaseSpeed         int64
	FuelConsumption   int64
}

// GetCargoCapacity returns ship cargo capacity
func (b BaseShip) GetCargoCapacity(techs IResearches, probeRaids bool, class CharacterClass, multiplier float64, bonuses LfBonuses) int64 {
	if b.GetID() == EspionageProbeID && !probeRaids {
		return 0
	}
	hyperspaceBonus := multiplier / 100
	lfBonus := float64(b.BaseCargoCapacity) * 0.01 * bonuses.ByShipID(b.GetID()).Cargo
	cargo := float64(b.BaseCargoCapacity) + lfBonus + float64(b.BaseCargoCapacity*techs.GetHyperspaceTechnology())*hyperspaceBonus
	if class == Collector && (b.ID == SmallCargoID || b.ID == LargeCargoID) {
		cargo += float64(b.BaseCargoCapacity) * 0.25
	}
	if class == General && b.ID == RecyclerID {
		cargo += float64(b.BaseCargoCapacity) * 0.2
	}
	return int64(math.Floor(cargo))
}

// GetFuelConsumption returns ship fuel consumption
func (b BaseShip) GetFuelConsumption(techs IResearches, fleetDeutSaveFactor float64, class CharacterClass, bonuses LfBonuses) int64 {
	if b.ID == EspionageProbeID {
		return 1
	}
	fuelConsumption := b.FuelConsumption
	if b.ID == SmallCargoID && techs.GetImpulseDrive() >= 5 {
		fuelConsumption *= 2
	} else if b.ID == RecyclerID && techs.GetHyperspaceDrive() >= 15 {
		fuelConsumption *= 3
	} else if b.ID == RecyclerID && techs.GetImpulseDrive() >= 17 {
		fuelConsumption *= 2
	}
	bonus := float64(fuelConsumption) * (1 - fleetDeutSaveFactor)
	bonus += float64(fuelConsumption) * 0.01 * bonuses.ByShipID(b.GetID()).Consumption
	if class == General {
		bonus += float64(fuelConsumption) * 0.25
	}
	return fuelConsumption - int64(bonus)
}

// GetSpeed returns speed of the ship
func (b BaseShip) GetSpeed(techs IResearches, class CharacterClass, bonuses LfBonuses) int64 {
	techDriveLvl := 0.0
	driveFactor := 0.2
	baseSpeed := float64(b.BaseSpeed)
	multiplier := int64(1)
	if b.ID == SmallCargoID && techs.GetImpulseDrive() >= 5 {
		baseSpeed = 10000
		techDriveLvl = float64(techs.GetImpulseDrive())
	} else if b.ID == BomberID && techs.GetHyperspaceDrive() >= 8 {
		baseSpeed = 5000
		techDriveLvl = float64(techs.GetHyperspaceDrive())
		driveFactor = 0.3
	} else if b.ID == RecyclerID && techs.GetHyperspaceDrive() >= 15 {
		techDriveLvl = float64(techs.GetHyperspaceDrive())
		multiplier = 3
		driveFactor = 0.3
	} else if b.ID == RecyclerID && techs.GetImpulseDrive() >= 17 {
		techDriveLvl = float64(techs.GetImpulseDrive())
		multiplier = 2
	} else if _, ok := b.Requirements[CombustionDriveID]; ok {
		techDriveLvl = float64(techs.GetCombustionDrive())
		driveFactor = 0.1
	} else if _, ok := b.Requirements[ImpulseDriveID]; ok {
		techDriveLvl = float64(techs.GetImpulseDrive())
	} else if _, ok := b.Requirements[HyperspaceDriveID]; ok {
		techDriveLvl = float64(techs.GetHyperspaceDrive())
		driveFactor = 0.3
	}
	bonus := 0.01 * baseSpeed * bonuses.ByShipID(b.GetID()).Speed
	speed := baseSpeed + (baseSpeed*driveFactor)*techDriveLvl + bonus
	if class == Collector && (b.ID == SmallCargoID || b.ID == LargeCargoID) {
		speed += baseSpeed
	} else if class == General && (b.ID == RecyclerID || b.ID.IsCombatShip()) && b.ID != DeathstarID {
		speed += baseSpeed
	}
	return int64(math.Round(speed)) * multiplier
}
