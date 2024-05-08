package v11_14_0

import (
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

// MODIFIED, TO BE TESTED
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
				report.Target.Type = ogame.PlanetType
				// TODO: implement moon type, the following extractor is not working 'cause the element always has the "planet" class
				if spanLink.Find("figure").HasClass("moon") {
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
	return msgs, 0, nil
}

// MODIFIED, TO BE TESTED
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
	if len(characterClassSplit) > 0 {
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
	// TODO: implement inactivity timer (there seems to be no info on the new detail page)

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
		// TODO: implement fleet levels (there seems to be no IDs on the elements)
	}

	defenses := doc.Find(".defenseSection")
	if defenses.Size() > 0 {
		report.HasDefensesInformation = true
		// TODO: implement defense levels (there seems to be no IDs on the elements)
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

// MODIFIED, TO BE TESTED
func extractCombatReportMessagesFromDoc(doc *goquery.Document) ([]ogame.CombatReportSummary, int64, error) {
	msgs := make([]ogame.CombatReportSummary, 0)
	doc.Find(".messagesHolder div.msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				report := ogame.CombatReportSummary{ID: id}
				report.Destination = v6.ExtractCoord(s.Find("div.msgHead a").Text())
				if s.Find("div.msgHead figure").HasClass("planet") {
					report.Destination.Type = ogame.PlanetType
				} else if s.Find("div.msgHead figure").HasClass("moon") {
					report.Destination.Type = ogame.MoonType
				} else {
					report.Destination.Type = ogame.PlanetType
				}
				apiKeyTitle := s.Find("span.icon_apikey").AttrOr("title", "")
				m := regexp.MustCompile(`'(cr-[^']+)'`).FindStringSubmatch(apiKeyTitle)
				if len(m) == 2 {
					report.APIKey = m[1]
				}
				report.Metal = utils.ParseInt(s.Find(".loot-row .metal .amount").Text())
				report.Crystal = utils.ParseInt(s.Find(".loot-row .crystal .amount").Text())
				report.Deuterium = utils.ParseInt(s.Find(".loot-row .deuterium .amount").Text())
				report.Food = utils.ParseInt(s.Find(".loot-row .food .amount").Text())
				report.DebrisField = utils.ParseInt(s.Find(".basicInfo .msg_ctn.msg_ctn3:nth-child(3)").AttrOr("title", "0"))
				m = regexp.MustCompile(`[\d.,]+[^\d]*([\d.,]+)`).FindStringSubmatch(s.Find(".basicInfo .msg_ctn.msg_ctn3:nth-child(2)").Eq(1).Text())
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
	return msgs, 0, nil
}

// MODIFIED, TO BE TESTED
func extractExpeditionMessagesFromDoc(doc *goquery.Document, location *time.Location) ([]ogame.ExpeditionMessage, int64, error) {
	msgs := make([]ogame.ExpeditionMessage, 0)
	doc.Find(".messagesHolder div.msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				msg := ogame.ExpeditionMessage{ID: id}
				msg.CreatedAt, _ = time.ParseInLocation("02.01.2006 15:04:05", s.Find("div.msgDate").Text(), location)
				msg.Coordinate = v6.ExtractCoord(s.Find("div.msgTitle a").Text())
				msg.Coordinate.Type = ogame.PlanetType
				msg.Content, _ = s.Find("div.msgContent").Html()
				msg.Content = strings.TrimSpace(msg.Content)
				msgs = append(msgs, msg)
			}
		}
	})
	return msgs, 0, nil
}
