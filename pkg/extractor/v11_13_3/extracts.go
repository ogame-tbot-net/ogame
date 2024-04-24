package v11_13_3

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/utils"
)

func extractLfBonusesFromDoc(doc *goquery.Document) (ogame.LfBonuses, error) {
	b := ogame.LfBonuses{}

	l := map[string]bool{
		"categoryResources":   true,
		"categoryShips":       true,
		"categoryCostAndTime": true,
		"categoryMisc":        false,
	}
	b.Ships = make(map[ogame.ID]ogame.ShipLfBonus)
	// extract resources bonus and ships cargo bonus for expeditions
	doc.Find("bonus-item-content[data-toggable-target^=category]").Each(func(_ int, s *goquery.Selection) {
		category, _ := s.Attr("data-toggable-target")
		if v, e := l[category]; e {
			if !v {
				return
			}
		} else {
			return
		}
		s.Find("inner-bonus-item-heading[data-toggable^=subcategory]").Each(func(_ int, g *goquery.Selection) {
			category, subcategory := extractCategories(g, category)
			if len(category) > 0 && len(subcategory) > 0 {
				b = assignBonusValue(g, b, category, subcategory)
			}
		})
	})

	return b, nil
}

// assign bonus value directly to LfBonuses struct
func assignBonusValue(g *goquery.Selection, b ogame.LfBonuses, category string, subcategory string) ogame.LfBonuses {
	switch category {
	case "Resources":
		switch subcategory {
		case "0":
			b.Production.Metal = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "1":
			b.Production.Crystal = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "2":
			b.Production.Deuterium = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "Expedition":
			b.Expeditions.Resources = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "ExpeditionShipsFound":
			b.Expeditions.Ships = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "ExpeditionSpeed":
			b.Expeditions.Speed = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		}
	case "Ships":
		b = extractShipStatBonusNew(g, b, subcategory)
	}
	return b
}

// extract raw bonus value for resources category
func extractRawResourcesBonusValue(g *goquery.Selection) string {
	return g.Find(".subCategoryBonus").Text()
}

// extract subcategories from attribute
func extractCategories(g *goquery.Selection, category string) (string, string) {
	c := strings.Replace(category, "category", "", 1)
	s, _ := g.Attr("data-toggable")
	v := "sub" + category
	return c, strings.Replace(s, v, "", 1)
}

// Extracts ogame id from a bonus item
func extractBonusID(g *goquery.Selection) ogame.LfBonusID {
	s, e := g.Attr("data-category")
	if !e {
		return 0
	}
	s = strings.Replace(s, "bonus-", "", 1)
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return ogame.LfBonusID(i)
}

// Extracts time reduction from all subitems
func extractTimeReductionBonus(s *goquery.Selection, l ogame.LfBonuses) ogame.LfBonuses {
	extractAllSubitems(s, func(id ogame.ID, bonuses *goquery.Selection) {
		if id.IsBuilding() {
			var tmp ogame.BaseLfBonus
			_, e := l.Buildings[id]
			if e {
				tmp = l.Buildings[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Buildings[id] = tmp
		} else if id.IsLfBuilding() {
			var tmp ogame.BaseLfBonus
			_, e := l.LfBuildings[id]
			if e {
				tmp = l.LfBuildings[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.LfBuildings[id] = tmp
		} else if id.IsTech() {
			var tmp ogame.BaseLfBonus
			_, e := l.Researches[id]
			if e {
				tmp = l.Researches[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Researches[id] = tmp
		} else if id.IsLfTech() {
			var tmp ogame.BaseLfBonus
			_, e := l.LfResearches[id]
			if e {
				tmp = l.LfResearches[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.LfResearches[id] = tmp
		} else if id.IsShip() {
			var tmp ogame.ShipLfBonus
			_, e := l.Ships[id]
			if e {
				tmp = l.Ships[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Ships[id] = tmp
		}
	})
	return l
}

// Extract bonus value from a string with percentage sign [ex: 1.056%]
func extractBonusFromStringPercentage(s string) float64 {
	v := strings.Replace(s, "%", "", 1)
	return extractBonusFromString(v)
}

// Extract bonus value from a string [ex: 1.056]
func extractBonusFromString(s string) float64 {
	v := strings.TrimSpace(s)
	v = strings.Replace(v, ",", ".", 1)
	b, _ := strconv.ParseFloat(v, 64)
	return utils.RoundThousandth(b)
}

// Extract bonus value from a row [ex: 1.056 / 30]
func extractBonusFromRow(r string) float64 {
	if strings.Contains(r, "/") {
		s := strings.Split(r, "/")
		return extractBonusFromString(s[0])
	}
	return 0
}

// Get all subitems in a category
func extractAllSubitems(s *goquery.Selection, clb func(ogame.ID, *goquery.Selection)) {
	s.Next().Find(".subItemContent").Each(func(_ int, g *goquery.Selection) {
		data, e := g.Find(".technology > button.details").Attr("data-target")
		if !e {
			return
		}
		reg := regexp.MustCompile(`technologyId=([0-9]+)`)
		raw := reg.FindStringSubmatch(data)
		if len(raw) != 2 {
			return
		}
		i, err := strconv.Atoi(raw[1])
		if err != nil {
			return
		}
		bonuses := g.Find(".innerSubItemHolder .innerSubItem > div:last-of-type")

		id := ogame.ID(i)
		if !id.IsValid() {
			return
		}

		clb(id, bonuses)
	})
}

// Extracts cost reduction from all subitems
func extractCostReductionBonus(s *goquery.Selection, l ogame.LfBonuses) ogame.LfBonuses {
	extractAllSubitems(s, func(id ogame.ID, bonuses *goquery.Selection) {
		if id.IsBuilding() {
			var tmp ogame.BaseLfBonus
			_, e := l.Buildings[id]
			if e {
				tmp = l.Buildings[id]
			}
			tmp.Cost = utils.RoundThousandth(tmp.Cost + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Buildings[id] = tmp
		} else if id.IsLfBuilding() {
			var tmp ogame.BaseLfBonus
			_, e := l.LfBuildings[id]
			if e {
				tmp = l.LfBuildings[id]
			}
			tmp.Cost = utils.RoundThousandth(tmp.Cost + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.LfBuildings[id] = tmp
		} else if id.IsTech() {
			var tmp ogame.BaseLfBonus
			_, e := l.Researches[id]
			if e {
				tmp = l.Researches[id]
			}
			tmp.Cost = utils.RoundThousandth(tmp.Cost + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Researches[id] = tmp
		} else if id.IsLfTech() {
			var tmp ogame.BaseLfBonus
			_, e := l.LfResearches[id]
			if e {
				tmp = l.LfResearches[id]
			}
			tmp.Cost = utils.RoundThousandth(tmp.Cost + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.LfResearches[id] = tmp
		}
	})
	return l
}

// Extracts ships consumption reduction
func extractShipConsumptionBonus(s *goquery.Selection, l ogame.LfBonuses) ogame.LfBonuses {
	extractAllSubitems(s, func(id ogame.ID, bonuses *goquery.Selection) {
		if id.IsShip() {
			_, e := l.Ships[id]
			if !e {
				l.Ships[id] = ogame.ShipLfBonus{}
			}
			tmp := l.Ships[id]
			tmp.Consumption = utils.RoundThousandth(tmp.Consumption + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Ships[id] = tmp
		}
	})
	return l
}

// Extracts ships stats fixed
func extractShipStatBonusNew(s *goquery.Selection, l ogame.LfBonuses, subcategory string) ogame.LfBonuses {
	i, err := strconv.Atoi(subcategory)
	if err != nil {
		return l
	}
	id := ogame.ID(i)
	if !id.IsValid() {
		return l
	}
	extractAllSubitemsNew(s, func(bonuses *goquery.Selection) {
		if id.IsShip() {
			_, e := l.Ships[id]
			if !e {
				l.Ships[id] = ogame.ShipLfBonus{}
			}
			tmp := l.Ships[id]
			tmp.Armour = tmp.Armour + extractBonusFromStringPercentage(bonuses.Children().Eq(0).Text())
			tmp.Shield = tmp.Shield + extractBonusFromStringPercentage(bonuses.Children().Eq(1).Text())
			tmp.Weapon = tmp.Weapon + extractBonusFromStringPercentage(bonuses.Children().Eq(2).Text())
			tmp.Speed = tmp.Speed + extractBonusFromStringPercentage(bonuses.Children().Eq(3).Text())
			tmp.Cargo = tmp.Cargo + extractBonusFromStringPercentage(bonuses.Children().Eq(4).Text())
			tmp.Consumption = tmp.Consumption + extractBonusFromStringPercentage(bonuses.Eq(5).Text())
			l.Ships[id] = tmp
		}
	})
	return l
}

// Get all bonuses from single ship
func extractAllSubitemsNew(s *goquery.Selection, clb func(*goquery.Selection)) {
	s.Find("bonus-items").Each(func(_ int, g *goquery.Selection) {
		bonuses := g
		clb(bonuses)
	})
}

// Extracts ships stats
func extractShipStatsBonus(s *goquery.Selection, l ogame.LfBonuses) ogame.LfBonuses {
	extractAllSubitems(s, func(id ogame.ID, bonuses *goquery.Selection) {
		if id.IsShip() {
			_, e := l.Ships[id]
			if !e {
				l.Ships[id] = ogame.ShipLfBonus{}
			}
			tmp := l.Ships[id]
			tmp.Armour = utils.RoundThousandth(tmp.Armour + extractBonusFromString(bonuses.Eq(0).Text()))
			tmp.Shield = utils.RoundThousandth(tmp.Shield + extractBonusFromString(bonuses.Eq(1).Text()))
			tmp.Weapon = utils.RoundThousandth(tmp.Weapon + extractBonusFromString(bonuses.Eq(2).Text()))
			tmp.Cargo = utils.RoundThousandth(tmp.Cargo + extractBonusFromString(bonuses.Eq(3).Text()))
			tmp.Speed = utils.RoundThousandth(tmp.Speed + extractBonusFromString(bonuses.Eq(4).Text()))
			l.Ships[id] = tmp
		} else if id.IsDefense() {
			_, e := l.Defenses[id]
			if !e {
				l.Defenses[id] = ogame.ShipLfBonus{}
			}
			tmp := l.Defenses[id]
			tmp.Armour = utils.RoundThousandth(tmp.Armour + extractBonusFromString(bonuses.Eq(0).Text()))
			tmp.Shield = utils.RoundThousandth(tmp.Shield + extractBonusFromString(bonuses.Eq(1).Text()))
			tmp.Weapon = utils.RoundThousandth(tmp.Weapon + extractBonusFromString(bonuses.Eq(2).Text()))
			l.Defenses[id] = tmp
		}
	})
	return l
}
