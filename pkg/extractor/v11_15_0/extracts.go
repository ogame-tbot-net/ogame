package v11_15_0

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	v6 "github.com/alaingilbert/ogame/pkg/extractor/v6"
	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/utils"
)

func extractCombatReportMessagesFromDoc(doc *goquery.Document) ([]ogame.CombatReportSummary, int64, error) {
	msgs := make([]ogame.CombatReportSummary, 0)
	doc.Find(".msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				rawMessageData := s.Find("div.rawMessageData")
				resultStr := rawMessageData.AttrOr("data-raw-result", "")
				var result struct {
					Loot struct {
						Percentage int64
						Resources  []struct {
							Resource string
							Amount   int64
						}
					}
				}
				_ = json.Unmarshal([]byte(resultStr), &result)

				report := ogame.CombatReportSummary{ID: id}
				report.Destination = v6.ExtractCoord(s.Find("div.msgHead a").Text())
				report.Destination.Type = ogame.PlanetType
				if s.Find("div.msgHead figure").HasClass("moon") {
					report.Destination.Type = ogame.MoonType
				}
				apiKeyTitle := s.Find("button.icon_apikey").AttrOr("title", "")
				m := regexp.MustCompile(`'(cr-[^']+)'`).FindStringSubmatch(apiKeyTitle)
				if len(m) == 2 {
					report.APIKey = m[1]
				}

				for _, resource := range result.Loot.Resources {
					res := resource.Resource
					if utils.InArr(res, []string{"deuter", "deuterij", "deutérium", "deuterium", "deuterio", "дейтерий", "deutério", "deuteriu", "デューテリウム", "重氫", "δευτέριο"}) {
						report.Deuterium = resource.Amount
					} else if utils.InArr(res, []string{"kristalli", "kristal", "cristal", "crystal", "krystal", "kryštály", "kryształ", "kristall", "krystall", "cristallo", "кристалл", "krystaly", "クリスタル", "晶體", "κρύσταλλο"}) {
						report.Crystal = resource.Amount
					} else if utils.InArr(res, []string{"metalli", "métal", "metal", "metall", "kov", "kovy", "металл", "metallo", "metaal", "メタル", "金屬", "μέταλλο"}) {
						report.Metal = resource.Amount
					}
				}

				debrisFieldTitle := s.Find("span.msg_content div.combatLeftSide span").Eq(2).AttrOr("title", "0")
				report.DebrisField = utils.ParseInt(debrisFieldTitle)
				resText := s.Find("span.msg_content div.combatLeftSide span").Eq(1).Text()
				m = regexp.MustCompile(`[\d.,]+[^\d]*([\d.,]+)`).FindStringSubmatch(resText)
				if len(m) == 2 {
					report.Loot = utils.ParseInt(m[1])
				}
				msgDate, _ := time.Parse("02.01.2006 15:04:05", s.Find("div.msgDate").Text())
				report.CreatedAt = msgDate

				link := s.Find("message-footer.msg_actions button.msgAttackBtn").AttrOr("onclick", "")
				m = regexp.MustCompile(`page=ingame&component=fleetdispatch&galaxy=(\d+)&system=(\d+)&position=(\d+)&type=(\d+)&`).FindStringSubmatch(link)
				if len(m) != 5 {
					return
				}
				galaxy := utils.DoParseI64(m[1])
				system := utils.DoParseI64(m[2])
				position := utils.DoParseI64(m[3])
				planetType := utils.DoParseI64(m[4])
				report.Origin = &ogame.Coordinate{Galaxy: galaxy, System: system, Position: position, Type: ogame.CelestialType(planetType)}
				if report.Origin.Equal(report.Destination) {
					report.Origin = nil
				}

				msgs = append(msgs, report)
			}
		}
	})
	return msgs, 1, nil
}

func extractEspionageReportMessageIDsFromDoc(doc *goquery.Document) ([]ogame.EspionageReportSummary, int64, error) {
	msgs := make([]ogame.EspionageReportSummary, 0)
	doc.Find(".msg").Each(func(i int, s *goquery.Selection) {
		rawData := s.Find("div.rawMessageData").First()
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				messageType := ogame.Report
				if s.Find("span.espionageDefText").Size() > 0 {
					messageType = ogame.Action
				}
				report := ogame.EspionageReportSummary{ID: id, Type: messageType}
				report.From = s.Find(".msgSender").Text()
				targetStr := rawData.AttrOr("data-raw-coordinates", "")
				report.Target, _ = ogame.ParseCoord(targetStr)
				report.Target.Type = ogame.PlanetType
				planetType := rawData.AttrOr("data-raw-targetplanettype", "1")
				if planetType == "3" {
					report.Target.Type = ogame.MoonType
				}
				if messageType == ogame.Report {
					s.Find(".lootPercentage").Each(func(i int, s *goquery.Selection) {
						if regexp.MustCompile(`%`).MatchString(s.Text()) {
							report.LootPercentage, _ = strconv.ParseFloat(regexp.MustCompile(`: (\d+)%`).FindStringSubmatch(s.Text())[1], 64)
							report.LootPercentage /= 100
						}
					})
				}
				msgs = append(msgs, report)
			}
		}
	})
	return msgs, 1, nil
}

func extractEspionageReportFromDoc(doc *goquery.Document, location *time.Location) (ogame.EspionageReport, error) {
	report := ogame.EspionageReport{}
	report.ID = utils.DoParseI64(doc.Find("div.detail_msg").AttrOr("data-msg-id", "0"))
	rawMessageData := doc.Find("div.rawMessageData").First()
	txt := rawMessageData.AttrOr("data-raw-coordinates", "")
	report.Coordinate = ogame.DoParseCoord(txt)
	if rawMessageData.AttrOr("data-raw-targetPlanetType", "1") == "1" {
		report.Coordinate.Type = ogame.PlanetType
	} else {
		report.Coordinate.Type = ogame.MoonType
	}
	messageType := ogame.Report
	//if doc.Find("span.espionageDefText").Size() > 0 {
	//	messageType = ogame.Action
	//}
	report.Type = messageType

	msgDateRaw := doc.Find("span.msg_date").Text()
	msgDate, _ := time.ParseInLocation("02.01.2006 15:04:05", msgDateRaw, location)
	report.Date = msgDate.In(time.Local)

	report.Username = strings.TrimSpace(rawMessageData.AttrOr("data-raw-playername", ""))

	characterClassJsonStr := strings.TrimSpace(rawMessageData.AttrOr("data-raw-characterclass", ""))
	var characterClassStruct struct{ ID int }
	_ = json.Unmarshal([]byte(characterClassJsonStr), &characterClassStruct)
	switch characterClassStruct.ID {
	case 1:
		report.CharacterClass = ogame.Collector
	case 2:
		report.CharacterClass = ogame.General
	case 3:
		report.CharacterClass = ogame.Discoverer
	default:
		report.CharacterClass = ogame.NoClass
	}

	allianceClassJsonStr := strings.TrimSpace(rawMessageData.AttrOr("data-raw-allianceclass", ""))
	var allianceClassStruct struct{ ID int }
	_ = json.Unmarshal([]byte(allianceClassJsonStr), &allianceClassStruct)
	switch allianceClassStruct.ID {
	case 1:
		report.AllianceClass = ogame.Warrior
	case 2:
		report.AllianceClass = ogame.Trader
	case 3:
		report.AllianceClass = ogame.Researcher
	default:
		report.AllianceClass = ogame.NoAllianceClass
	}

	// Bandit, Starlord
	banditstarlord := doc.Find("span.honorRank").First()
	if banditstarlord.HasClass("honorRank") {
		report.IsBandit = banditstarlord.HasClass("rank_bandit1") || banditstarlord.HasClass("rank_bandit2") || banditstarlord.HasClass("rank_bandit3")
		report.IsStarlord = banditstarlord.HasClass("rank_starlord1") || banditstarlord.HasClass("rank_starlord2") || banditstarlord.HasClass("rank_starlord3")
	}

	report.HonorableTarget = doc.Find("span.status_abbr_honorableTarget").Length() > 0

	// IsInactive, IsLongInactive
	inactive := doc.Find("div.playerInfo").First().Find("span")
	if inactive.HasClass("status_abbr_longinactive") {
		report.IsInactive = true
		report.IsLongInactive = true
	} else if inactive.HasClass("status_abbr_inactive") {
		report.IsInactive = true
	}

	// APIKey
	apikey, _ := doc.Find("button.icon_apikey").Attr("title")
	apiDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(apikey))
	report.APIKey = apiDoc.Find("input").First().AttrOr("value", "")

	// Inactivity timer
	report.LastActivity = utils.ParseInt(rawMessageData.AttrOr("data-raw-activity", "-1"))
	if report.LastActivity == -1 {
		report.LastActivity = 0
	}

	// CounterEspionage
	report.CounterEspionage = utils.ParseInt(rawMessageData.AttrOr("data-raw-counterespionagechance", "0"))

	report.Metal = utils.DoParseI64(rawMessageData.AttrOr("data-raw-metal", "0"))
	report.Crystal = utils.DoParseI64(rawMessageData.AttrOr("data-raw-crystal", "0"))
	report.Deuterium = utils.DoParseI64(rawMessageData.AttrOr("data-raw-deuterium", "0"))
	report.Food = utils.DoParseI64(rawMessageData.AttrOr("data-raw-food", "0"))
	report.Population = utils.DoParseI64(rawMessageData.AttrOr("data-raw-population", "0"))
	report.Energy = utils.ParseInt(doc.Find("resource-icon.energy").Next().Text())

	report.HasBuildingsInformation = rawMessageData.AttrOr("data-raw-hiddenbuildings", "1") == ""
	if report.HasBuildingsInformation {
		buildingsStr := rawMessageData.AttrOr("data-raw-buildings", "{}")
		var buildingsStruct struct {
			MetalMine            *int64 `json:"1"`
			CrystalMine          *int64 `json:"2"`
			DeuteriumSynthesizer *int64 `json:"3"`
			SolarPlant           *int64 `json:"4"`
			FusionReactor        *int64 `json:"12"`
			MetalStorage         *int64 `json:"22"`
			CrystalStorage       *int64 `json:"23"`
			DeuteriumTank        *int64 `json:"24"`
			AllianceDepot        *int64 `json:"34"`
			RoboticsFactory      *int64 `json:"14"`
			Shipyard             *int64 `json:"21"`
			ResearchLab          *int64 `json:"31"`
			MissileSilo          *int64 `json:"44"`
			NaniteFactory        *int64 `json:"15"`
			Terraformer          *int64 `json:"33"`
			SpaceDock            *int64 `json:"36"`
			LunarBase            *int64 `json:"41"`
			SensorPhalanx        *int64 `json:"42"`
			JumpGate             *int64 `json:"43"`
		}
		_ = json.Unmarshal([]byte(buildingsStr), &buildingsStruct)
		report.MetalMine = buildingsStruct.MetalMine
		report.CrystalMine = buildingsStruct.CrystalMine
		report.DeuteriumSynthesizer = buildingsStruct.DeuteriumSynthesizer
		report.SolarPlant = buildingsStruct.SolarPlant
		report.FusionReactor = buildingsStruct.FusionReactor
		report.MetalStorage = buildingsStruct.MetalStorage
		report.CrystalStorage = buildingsStruct.CrystalStorage
		report.DeuteriumTank = buildingsStruct.DeuteriumTank
		report.AllianceDepot = buildingsStruct.AllianceDepot
		report.RoboticsFactory = buildingsStruct.RoboticsFactory
		report.Shipyard = buildingsStruct.Shipyard
		report.ResearchLab = buildingsStruct.ResearchLab
		report.MissileSilo = buildingsStruct.MissileSilo
		report.NaniteFactory = buildingsStruct.NaniteFactory
		report.Terraformer = buildingsStruct.Terraformer
		report.SpaceDock = buildingsStruct.SpaceDock
		report.LunarBase = buildingsStruct.LunarBase
		report.SensorPhalanx = buildingsStruct.SensorPhalanx
		report.JumpGate = buildingsStruct.JumpGate
	}
	report.HasResearchesInformation = rawMessageData.AttrOr("data-raw-hiddenresearch", "1") == ""
	if report.HasResearchesInformation {
		researchStr := rawMessageData.AttrOr("data-raw-research", "{}")
		var researchStruct struct {
			EspionageTechnology          *int64 `json:"106"`
			ComputerTechnology           *int64 `json:"108"`
			WeaponsTechnology            *int64 `json:"109"`
			ShieldingTechnology          *int64 `json:"110"`
			ArmourTechnology             *int64 `json:"111"`
			EnergyTechnology             *int64 `json:"113"`
			HyperspaceTechnology         *int64 `json:"114"`
			CombustionDrive              *int64 `json:"115"`
			ImpulseDrive                 *int64 `json:"117"`
			HyperspaceDrive              *int64 `json:"118"`
			LaserTechnology              *int64 `json:"120"`
			IonTechnology                *int64 `json:"121"`
			PlasmaTechnology             *int64 `json:"122"`
			IntergalacticResearchNetwork *int64 `json:"123"`
			Astrophysics                 *int64 `json:"124"`
			GravitonTechnology           *int64 `json:"199"`
		}
		_ = json.Unmarshal([]byte(researchStr), &researchStruct)
		report.EspionageTechnology = researchStruct.EspionageTechnology
		report.ComputerTechnology = researchStruct.ComputerTechnology
		report.WeaponsTechnology = researchStruct.WeaponsTechnology
		report.ShieldingTechnology = researchStruct.ShieldingTechnology
		report.ArmourTechnology = researchStruct.ArmourTechnology
		report.EnergyTechnology = researchStruct.EnergyTechnology
		report.HyperspaceTechnology = researchStruct.HyperspaceTechnology
		report.CombustionDrive = researchStruct.CombustionDrive
		report.ImpulseDrive = researchStruct.ImpulseDrive
		report.HyperspaceDrive = researchStruct.HyperspaceDrive
		report.LaserTechnology = researchStruct.LaserTechnology
		report.IonTechnology = researchStruct.IonTechnology
		report.PlasmaTechnology = researchStruct.PlasmaTechnology
		report.IntergalacticResearchNetwork = researchStruct.IntergalacticResearchNetwork
		report.Astrophysics = researchStruct.Astrophysics
		report.GravitonTechnology = researchStruct.GravitonTechnology
	}

	report.HasFleetInformation = rawMessageData.AttrOr("data-raw-hiddenships", "1") == ""
	if report.HasFleetInformation {
		fleetStr := rawMessageData.AttrOr("data-raw-fleet", "{}")
		var fleetStruct struct {
			SmallCargo     *int64 `json:"202"`
			LargeCargo     *int64 `json:"203"`
			LightFighter   *int64 `json:"204"`
			HeavyFighter   *int64 `json:"205"`
			Cruiser        *int64 `json:"206"`
			Battleship     *int64 `json:"207"`
			ColonyShip     *int64 `json:"208"`
			Recycler       *int64 `json:"209"`
			EspionageProbe *int64 `json:"210"`
			Bomber         *int64 `json:"211"`
			SolarSatellite *int64 `json:"212"`
			Destroyer      *int64 `json:"213"`
			Deathstar      *int64 `json:"214"`
			Battlecruiser  *int64 `json:"215"`
			Crawler        *int64 `json:"217"`
			Reaper         *int64 `json:"218"`
			Pathfinder     *int64 `json:"219"`
		}
		_ = json.Unmarshal([]byte(fleetStr), &fleetStruct)
		report.SmallCargo = fleetStruct.SmallCargo
		report.LargeCargo = fleetStruct.LargeCargo
		report.LightFighter = fleetStruct.LightFighter
		report.HeavyFighter = fleetStruct.HeavyFighter
		report.Cruiser = fleetStruct.Cruiser
		report.Battleship = fleetStruct.Battleship
		report.ColonyShip = fleetStruct.ColonyShip
		report.Recycler = fleetStruct.Recycler
		report.EspionageProbe = fleetStruct.EspionageProbe
		report.Bomber = fleetStruct.Bomber
		report.SolarSatellite = fleetStruct.SolarSatellite
		report.Destroyer = fleetStruct.Destroyer
		report.Deathstar = fleetStruct.Deathstar
		report.Battlecruiser = fleetStruct.Battlecruiser
		report.Crawler = fleetStruct.Crawler
		report.Reaper = fleetStruct.Reaper
		report.Pathfinder = fleetStruct.Pathfinder
	}

	report.HasDefensesInformation = rawMessageData.AttrOr("data-raw-hiddendef", "1") == ""
	if report.HasDefensesInformation {
		defStr := rawMessageData.AttrOr("data-raw-defense", "{}")
		var defStruct struct {
			RocketLauncher         *int64 `json:"401"`
			LightLaser             *int64 `json:"402"`
			HeavyLaser             *int64 `json:"403"`
			GaussCannon            *int64 `json:"404"`
			IonCannon              *int64 `json:"405"`
			PlasmaTurret           *int64 `json:"406"`
			SmallShieldDome        *int64 `json:"407"`
			LargeShieldDome        *int64 `json:"408"`
			AntiBallisticMissiles  *int64 `json:"502"`
			InterplanetaryMissiles *int64 `json:"503"`
		}
		_ = json.Unmarshal([]byte(defStr), &defStruct)
		report.RocketLauncher = defStruct.RocketLauncher
		report.LightLaser = defStruct.LightLaser
		report.HeavyLaser = defStruct.HeavyLaser
		report.GaussCannon = defStruct.GaussCannon
		report.IonCannon = defStruct.IonCannon
		report.PlasmaTurret = defStruct.PlasmaTurret
		report.SmallShieldDome = defStruct.SmallShieldDome
		report.LargeShieldDome = defStruct.LargeShieldDome
		report.AntiBallisticMissiles = defStruct.AntiBallisticMissiles
		report.InterplanetaryMissiles = defStruct.InterplanetaryMissiles
	}

	return report, nil
}

func extractExpeditionMessagesFromDoc(doc *goquery.Document, location *time.Location) ([]ogame.ExpeditionMessage, int64, error) {
	msgs := make([]ogame.ExpeditionMessage, 0)
	doc.Find(".msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				msg := ogame.ExpeditionMessage{ID: id}
				msg.CreatedAt, _ = time.ParseInLocation("02.01.2006 15:04:05", s.Find(".msgDate").Text(), location)
				msg.Coordinate = v6.ExtractCoord(s.Find(".msgTitle a").Text())
				msg.Coordinate.Type = ogame.PlanetType
				msg.Content, _ = s.Find("div.msgContent").Html()
				msg.Content = strings.TrimSpace(msg.Content)

				var resStruct struct {
					Metal      int64 `json:"metal"`
					Crystal    int64 `json:"crystal"`
					Deuterium  int64 `json:"deuterium"`
					Darkmatter int64 `json:"darkMatter"`
				}
				resGained := s.Find("div.rawMessageData").AttrOr("data-raw-resourcesgained", "{}")
				_ = json.Unmarshal([]byte(resGained), &resStruct)

				msg.Resources.Metal = resStruct.Metal
				msg.Resources.Crystal = resStruct.Crystal
				msg.Resources.Deuterium = resStruct.Deuterium
				msg.Resources.Darkmatter = resStruct.Darkmatter

				type ShipInfo struct {
					Amount int64  `json:"amount"`
					Name   string `json:"name"`
				}
				var msgDataStruct struct {
					SmallCargo     ShipInfo `json:"202"`
					LargeCargo     ShipInfo `json:"203"`
					LightFighter   ShipInfo `json:"204"`
					HeavyFighter   ShipInfo `json:"205"`
					Cruiser        ShipInfo `json:"206"`
					Battleship     ShipInfo `json:"207"`
					EspionageProbe ShipInfo `json:"210"`
					Bomber         ShipInfo `json:"211"`
					Destroyer      ShipInfo `json:"213"`
					Battlecruiser  ShipInfo `json:"215"`
					Reaper         ShipInfo `json:"218"`
					Pathfinder     ShipInfo `json:"219"`
				}
				techGained := s.Find("div.rawMessageData").AttrOr("data-raw-technologiesgained", "{}")
				_ = json.Unmarshal([]byte(techGained), &msgDataStruct)
				msg.Ships.SmallCargo = msgDataStruct.SmallCargo.Amount
				msg.Ships.LargeCargo = msgDataStruct.LargeCargo.Amount
				msg.Ships.LightFighter = msgDataStruct.LightFighter.Amount
				msg.Ships.HeavyFighter = msgDataStruct.HeavyFighter.Amount
				msg.Ships.Cruiser = msgDataStruct.Cruiser.Amount
				msg.Ships.Battleship = msgDataStruct.Battleship.Amount
				msg.Ships.EspionageProbe = msgDataStruct.EspionageProbe.Amount
				msg.Ships.Bomber = msgDataStruct.Bomber.Amount
				msg.Ships.Destroyer = msgDataStruct.Destroyer.Amount
				msg.Ships.Battlecruiser = msgDataStruct.Battlecruiser.Amount
				msg.Ships.Reaper = msgDataStruct.Reaper.Amount
				msg.Ships.Pathfinder = msgDataStruct.Pathfinder.Amount

				msgs = append(msgs, msg)
			}
		}
	})
	return msgs, 1, nil
}

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
	case "CostAndTime":
		b = extractCostReductionBonus(g, b)
		b = extractTimeReductionBonus(g, b)
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

func extractAllianceClassFromDoc(doc *goquery.Document) ogame.AllianceClass {
	allianceClass := ogame.NoAllianceClass
	el := doc.Find("div.allianceclass").First()
	if el.HasClass("warrior") {
		allianceClass = ogame.Warrior
	} else if el.HasClass("trader") {
		allianceClass = ogame.Trader
	} else if el.HasClass("explorer") {
		allianceClass = ogame.Researcher
	}
	return allianceClass
}
