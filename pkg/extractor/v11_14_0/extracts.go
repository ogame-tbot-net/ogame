package v11_14_0

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	v6 "github.com/alaingilbert/ogame/pkg/extractor/v6"
	v71 "github.com/alaingilbert/ogame/pkg/extractor/v71"
	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/utils"
)

func ExtractCoord(v string) (coord ogame.Coordinate) {
	coordRgx := regexp.MustCompile(`\[(\d+):(\d+):(\d+)]`)
	m := coordRgx.FindStringSubmatch(v)
	if len(m) == 4 {
		coord.Galaxy = utils.DoParseI64(m[1])
		coord.System = utils.DoParseI64(m[2])
		coord.Position = utils.DoParseI64(m[3])
	}
	return
}

// OK, PLANET TYPE RETURNS ALWAYS PLANET
func extractEspionageReportMessageIDsFromDoc(doc *goquery.Document) ([]ogame.EspionageReportSummary, int64, error) {
	msgs := make([]ogame.EspionageReportSummary, 0)
	doc.Find(".msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				messageType := ogame.Report
				if s.Find("span.espionageDefText").Size() > 0 {
					messageType = ogame.Action
				}
				report := ogame.EspionageReportSummary{ID: id, Type: messageType}
				report.From = s.Find(".msgSender").Text()
				spanLink := s.Find(".msgTitle a")
				targetStr := spanLink.Text()
				report.Target = ExtractCoord(targetStr)

				// TODO: implement moon type, the following extractor is not working 'cause the element always has the "planet" class
				report.Target.Type = ogame.PlanetType

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
	return msgs, 0, nil
}

// OK, MISSING A COUPLE OF NOT VERY IMPORTANT THINGS
func extractEspionageReportFromDoc(doc *goquery.Document, location *time.Location) (ogame.EspionageReport, error) {
	report := ogame.EspionageReport{}
	report.ID = utils.DoParseI64(doc.Find("div.detail_msg").AttrOr("data-msg-id", "0"))
	spanLink := doc.Find("span.msg_title a").First()
	txt := spanLink.Text()
	report.Coordinate = ExtractCoord(txt)
	figure := spanLink.Find("figure").First()
	if figure.HasClass("planet") {
		report.Coordinate.Type = ogame.PlanetType
	} else if figure.HasClass("moon") {
		report.Coordinate.Type = ogame.MoonType
	}
	messageType := ogame.Report
	if doc.Find("span.espionageDefText").Size() > 0 {
		messageType = ogame.Action
	}
	report.Type = messageType
	msgDateRaw := doc.Find("span.msg_date").Text()
	msgDate, _ := time.ParseInLocation("02.01.2006 15:04:05", msgDateRaw, location)
	report.Date = msgDate.In(time.Local)

	username := doc.Find(".playerInfo .playerName").Find("span").First().Text()
	report.Username = strings.TrimSpace(username)
	characterClassStr := doc.Find(".playerInfo .characterClassInfo").First().Text()
	characterClassStr = strings.TrimSpace(characterClassStr)
	characterClassSplit := strings.Split(characterClassStr, ":")
	if len(characterClassSplit) > 1 {
		report.CharacterClass = v71.GetCharacterClass(strings.TrimSpace(characterClassSplit[1]))
	}

	report.AllianceClass = ogame.NoAllianceClass
	allianceClassStr := doc.Find(".playerInfo .allianceClassInfo").Text()
	allianceClassStr = strings.TrimSpace(allianceClassStr)
	allianceClassSplit := strings.Split(allianceClassStr, ":")
	if len(allianceClassSplit) > 0 {
		// TODO: implement an extractor for alliance class
		report.AllianceClass = ogame.NoAllianceClass
	}

	// Bandit, Starlord
	banditstarlord := doc.Find(".playerInfo .playerName").First().Find("span").First().Find("span").First()
	if banditstarlord.HasClass("honorRank") {
		report.IsBandit = banditstarlord.HasClass("rank_bandit1") || banditstarlord.HasClass("rank_bandit2") || banditstarlord.HasClass("rank_bandit3")
		report.IsStarlord = banditstarlord.HasClass("rank_starlord1") || banditstarlord.HasClass("rank_starlord2") || banditstarlord.HasClass("rank_starlord3")
	}

	honorableFound := doc.Find(".playerInfo .playerName").First().Find("span.status_abbr_honorableTarget")
	report.HonorableTarget = honorableFound.Length() > 0

	// IsInactive, IsLongInactive
	inactive := doc.Find(".playerInfo .playerName").First().Find("span")
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
	report.LastActivity = utils.DoParseI64(doc.Find(".detailsActivity span").Text())

	// CounterEspionage
	ceTxt := doc.Find(".miscInfo .counterInfo").Text()
	m1 := regexp.MustCompile(`(\d+)%`).FindStringSubmatch(ceTxt)
	if len(m1) == 2 {
		report.CounterEspionage = utils.DoParseI64(m1[1])
	}

	hasError := false
	report.HasFleetInformation = false
	report.HasDefensesInformation = false
	report.HasBuildingsInformation = false
	report.HasResearchesInformation = false

	res := doc.Find(".resourceLootInfo .loot-row")
	if res.Size() > 0 {
		report.Metal = utils.ParseInt(res.Find("resource-icon.metal .amount").First().AttrOr("title", "0"))
		report.Crystal = utils.ParseInt(res.Find("resource-icon.crystal .amount").First().AttrOr("title", "0"))
		report.Deuterium = utils.ParseInt(res.Find("resource-icon.deuterium .amount").First().AttrOr("title", "0"))
		report.Food = utils.ParseInt(res.Find("resource-icon.food .amount").First().AttrOr("title", "0"))
		report.Population = utils.ParseInt(res.Find("resource-icon.population .amount").First().AttrOr("title", "0"))
		report.Energy = utils.ParseInt(res.Find(".energy .amount").First().Text())
	}

	fleet := doc.Find(".fleetSection")
	if fleet.Size() > 0 {
		report.HasFleetInformation = true
		lightFighter := utils.ParseInt(fleet.Find(".loot-item technology-icon[fighterlight]").Siblings().First().Find(".loot-amount").Text())
		report.LightFighter = &lightFighter
		heavyFighter := utils.ParseInt(fleet.Find(".loot-item technology-icon[fighterheavy]").Siblings().First().Find(".loot-amount").Text())
		report.HeavyFighter = &heavyFighter
		cruiser := utils.ParseInt(fleet.Find(".loot-item technology-icon[cruiser]").Siblings().First().Find(".loot-amount").Text())
		report.Cruiser = &cruiser
		battleship := utils.ParseInt(fleet.Find(".loot-item technology-icon[battleship]").Siblings().First().Find(".loot-amount").Text())
		report.Battleship = &battleship
		battlecruiser := utils.ParseInt(fleet.Find(".loot-item technology-icon[interceptor]").Siblings().First().Find(".loot-amount").Text())
		report.Battlecruiser = &battlecruiser
		bomber := utils.ParseInt(fleet.Find(".loot-item technology-icon[bomber]").Siblings().First().Find(".loot-amount").Text())
		report.Bomber = &bomber
		destroyer := utils.ParseInt(fleet.Find(".loot-item technology-icon[destroyer]").Siblings().First().Find(".loot-amount").Text())
		report.Destroyer = &destroyer
		deathstar := utils.ParseInt(fleet.Find(".loot-item technology-icon[deathstar]").Siblings().First().Find(".loot-amount").Text())
		report.Deathstar = &deathstar
		reaper := utils.ParseInt(fleet.Find(".loot-item technology-icon[reaper]").Siblings().First().Find(".loot-amount").Text())
		report.Reaper = &reaper
		pathfinder := utils.ParseInt(fleet.Find(".loot-item technology-icon[explorer]").Siblings().First().Find(".loot-amount").Text())
		report.Pathfinder = &pathfinder
		smallCargo := utils.ParseInt(fleet.Find(".loot-item technology-icon[transportersmall]").Siblings().First().Find(".loot-amount").Text())
		report.SmallCargo = &smallCargo
		largeCargo := utils.ParseInt(fleet.Find(".loot-item technology-icon[transporterlarge]").Siblings().First().Find(".loot-amount").Text())
		report.LargeCargo = &largeCargo
		colonyShip := utils.ParseInt(fleet.Find(".loot-item technology-icon[colonyship]").Siblings().First().Find(".loot-amount").Text())
		report.ColonyShip = &colonyShip
		recycler := utils.ParseInt(fleet.Find(".loot-item technology-icon[recycler]").Siblings().First().Find(".loot-amount").Text())
		report.Recycler = &recycler
		espionageProbe := utils.ParseInt(fleet.Find(".loot-item technology-icon[espionageprobe]").Siblings().First().Find(".loot-amount").Text())
		report.EspionageProbe = &espionageProbe
		solarSatellite := utils.ParseInt(fleet.Find(".loot-item technology-icon[solarsatellite]").Siblings().First().Find(".loot-amount").Text())
		report.SolarSatellite = &solarSatellite
		crawler := utils.ParseInt(fleet.Find(".loot-item technology-icon[resbuggy]").Siblings().First().Find(".loot-amount").Text())
		report.Crawler = &crawler
	}

	defenses := doc.Find(".defenseSection")
	if defenses.Size() > 0 {
		report.HasDefensesInformation = true
		rocketLauncher := utils.ParseInt(defenses.Find(".loot-item technology-icon[rocketlauncher]").Siblings().First().Find(".loot-amount").Text())
		report.RocketLauncher = &rocketLauncher
		lightLaser := utils.ParseInt(defenses.Find(".loot-item technology-icon[lasercannonlight]").Siblings().First().Find(".loot-amount").Text())
		report.LightLaser = &lightLaser
		heavyLaser := utils.ParseInt(defenses.Find(".loot-item technology-icon[lasercannonheavy]").Siblings().First().Find(".loot-amount").Text())
		report.HeavyLaser = &heavyLaser
		gaussCannon := utils.ParseInt(defenses.Find(".loot-item technology-icon[gausscannon]").Siblings().First().Find(".loot-amount").Text())
		report.GaussCannon = &gaussCannon
		ionCannon := utils.ParseInt(defenses.Find(".loot-item technology-icon[ioncannon]").Siblings().First().Find(".loot-amount").Text())
		report.IonCannon = &ionCannon
		plasmaTurret := utils.ParseInt(defenses.Find(".loot-item technology-icon[plasmacannon]").Siblings().First().Find(".loot-amount").Text())
		report.PlasmaTurret = &plasmaTurret
		smallShieldDome := utils.ParseInt(defenses.Find(".loot-item technology-icon[shielddomesmall]").Siblings().First().Find(".loot-amount").Text())
		report.SmallShieldDome = &smallShieldDome
		largeShieldDome := utils.ParseInt(defenses.Find(".loot-item technology-icon[shielddomelarge]").Siblings().First().Find(".loot-amount").Text())
		report.LargeShieldDome = &largeShieldDome
		antiBallisticMissiles := utils.ParseInt(defenses.Find(".loot-item technology-icon[missileinterceptor]").Siblings().First().Find(".loot-amount").Text())
		report.AntiBallisticMissiles = &antiBallisticMissiles
		interplanetaryMissiles := utils.ParseInt(defenses.Find(".loot-item technology-icon[missileinterplanetary]").Siblings().First().Find(".loot-amount").Text())
		report.InterplanetaryMissiles = &interplanetaryMissiles
	}

	buildings := doc.Find(".buildingsSection")
	if buildings.Size() > 0 {
		report.HasBuildingsInformation = true
		// TODO: implement building levels (there seems to be no IDs on the elements)
	}

	researches := doc.Find(".researchSection")
	if researches.Size() > 0 {
		report.HasResearchesInformation = true
		// TODO: implement research levels (there seems to be no IDs on the elements)
	}

	if hasError {
		return report, ogame.ErrDeactivateHidePictures
	}
	return report, nil
}

// OK, MISSING DETAILED RESOURCES INFO
func extractCombatReportMessagesFromDoc(doc *goquery.Document) ([]ogame.CombatReportSummary, int64, error) {
	msgs := make([]ogame.CombatReportSummary, 0)
	doc.Find(".msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				report := ogame.CombatReportSummary{ID: id}
				report.Destination = v6.ExtractCoord(s.Find("div.msgHead a").Text())
				report.Destination.Type = ogame.PlanetType
				if s.Find("div.msgHead figure").HasClass("moon") {
					report.Destination.Type = ogame.MoonType
				}

				attackerName := s.Find(".basicInfo .msg_ctn.msg_ctn2").Text()
				m := regexp.MustCompile(`\([^']+\)`).FindStringSubmatch(attackerName)
				if len(m) == 1 {
					report.AttackerName = m[0][1 : len(m[0])-1]
				}

				defenderName := s.Find(".miscInfo .msg_ctn.msg_ctn2").Text()
				m = regexp.MustCompile(`\([^']+\)`).FindStringSubmatch(defenderName)
				if len(m) == 1 {
					report.DefenderName = m[0][1 : len(m[0])-1]
				}

				apiKeyTitle := s.Find(".icon_apikey").AttrOr("title", "")
				m = regexp.MustCompile(`'(cr-[^']+)'`).FindStringSubmatch(apiKeyTitle)
				if len(m) == 2 {
					report.APIKey = m[1]
				}

				// There are no more detailed infos about looted resources in the combat report summary
				rel := s.Find(".basicInfo .msg_ctn.msg_ctn3:nth-child(2)").Text()
				m = regexp.MustCompile(`[\d.,]+[^\d]*([\d.,]+)`).FindStringSubmatch(rel)
				if len(m) == 2 {
					report.Loot = utils.ParseInt(m[1])
				}
				m = regexp.MustCompile(`: ([\d.,]+)`).FindStringSubmatch(rel)
				if len(m) == 2 {
					report.Resources = utils.ParseInt(m[1][0 : len(m[1])-1])
				}

				report.DebrisField = utils.ParseInt(s.Find(".basicInfo .msg_ctn.msg_ctn3:nth-child(3)").AttrOr("title", "0"))

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
	return msgs, 0, nil
}

// OK
func extractExpeditionMessagesFromDoc(doc *goquery.Document, location *time.Location) ([]ogame.ExpeditionMessage, int64, error) {
	msgs := make([]ogame.ExpeditionMessage, 0)
	doc.Find(".msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				msg := ogame.ExpeditionMessage{ID: id}

				msg.CreatedAt, _ = time.ParseInLocation("02.01.2006 15:04:05", s.Find("div.msgDate").Text(), location)

				msg.Coordinate = v6.ExtractCoord(s.Find("div.msgTitle a").Text())
				msg.Coordinate.Type = ogame.PlanetType

				msg.Content, _ = s.Find("div.msgContent").Html()
				msg.Content = strings.TrimSpace(msg.Content)

				msg.Resources.Metal = utils.ParseInt(s.Find("resource-icon.metal .amount").First().AttrOr("title", "0"))
				msg.Resources.Crystal = utils.ParseInt(s.Find("resource-icon.crystal .amount").First().AttrOr("title", "0"))
				msg.Resources.Deuterium = utils.ParseInt(s.Find("resource-icon.deuterium .amount").First().AttrOr("title", "0"))
				msg.Resources.Darkmatter = utils.ParseInt(s.Find("resource-icon.darkmatter .amount").First().Text())

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
				json.Unmarshal([]byte(techGained), &msgDataStruct)
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
	return msgs, 0, nil
}
