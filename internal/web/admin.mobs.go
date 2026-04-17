package web

import (
	"fmt"
	htemplate "html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

type adminMobOption struct {
	mobs.Mob
	DataContent htemplate.HTMLAttr
}

func mobsIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := htemplate.New("index.html").Funcs(funcMap).ParseFiles(configs.GetFilePathsConfig().AdminHtml.String()+"/_header.html", configs.GetFilePathsConfig().AdminHtml.String()+"/mobs/index.html", configs.GetFilePathsConfig().AdminHtml.String()+"/_footer.html")
	if err != nil {
		mudlog.Error("HTML Template", "error", err)
	}

	allMobs := mobs.GetAllMobInfo()
	sort.SliceStable(allMobs, func(i, j int) bool {
		return allMobs[i].MobId < allMobs[j].MobId
	})

	mobOptions := make([]adminMobOption, 0, len(allMobs))
	for _, mobInfo := range allMobs {
		mobOptions = append(mobOptions, adminMobOption{
			Mob:         mobInfo,
			DataContent: adminMobDataContent(mobInfo),
		})
	}

	mobIndexData := struct {
		Mobs []adminMobOption
	}{
		mobOptions,
	}

	if err := tmpl.Execute(w, mobIndexData); err != nil {
		mudlog.Error("HTML Execute", "error", err)
	}

}

func adminMobDataContent(mobInfo mobs.Mob) htemplate.HTMLAttr {
	var markup strings.Builder

	fmt.Fprintf(&markup, "<span class='badge badge-secondary'>%d</span> ", mobInfo.MobId)

	if len(mobInfo.QuestFlags) > 0 {
		markup.WriteString("<span class='text-warning'>&#x2605;</span> ")
	}

	fmt.Fprintf(&markup, "<span class='font-weight-bold'>%s</span>", adminPickerDataText(mobInfo.Character.Name))

	if len(mobInfo.Character.Shop) > 0 {
		markup.WriteString(" <span class='badge badge-pill badge-warning'>shop</span>")
	}

	if mobInfo.GetScript() != "" {
		markup.WriteString(" <span class='badge badge-pill badge-info'>Script</span>")
	}

	return htemplate.HTMLAttr(`data-content="` + markup.String() + `"`)
}

func mobData(w http.ResponseWriter, r *http.Request) {

	tmpl, err := htemplate.New("mob.data.html").Funcs(funcMap).ParseFiles(configs.GetFilePathsConfig().AdminHtml.String() + "/mobs/mob.data.html")
	if err != nil {
		mudlog.Error("HTML Template", "error", err)
	}

	urlVals := r.URL.Query()

	mobIdInt, _ := strconv.Atoi(urlVals.Get(`mobid`))

	mobInfo := mobs.GetMobSpec(mobs.MobId(mobIdInt))
	if mobInfo == nil {
		mobInfo = &mobs.Mob{}
	}

	mobGroupSet := map[string]struct{}{}
	allMobGroups := []string{}
	for _, m := range mobs.GetAllMobInfo() {

		for _, groupName := range m.Groups {
			if _, ok := mobGroupSet[groupName]; !ok {
				allMobGroups = append(allMobGroups, groupName)
				mobGroupSet[groupName] = struct{}{}
			}
		}

	}

	allRaces := races.GetRaces()
	sort.SliceStable(allRaces, func(i, j int) bool {
		return allRaces[i].RaceId < allRaces[j].RaceId
	})

	allZoneNames := rooms.GetAllZoneNames()
	sort.SliceStable(allZoneNames, func(i, j int) bool {
		return allZoneNames[i] < allZoneNames[j]
	})

	activityLevels := []int{}
	for i := 1; i < 101; i++ {
		activityLevels = append(activityLevels, i)
	}

	dropChances := []int{}
	for i := 0; i < 101; i++ {
		dropChances = append(dropChances, i)
	}

	buffSpecs := []buffs.BuffSpec{}
	for _, buffId := range buffs.GetAllBuffIds() {
		if b := buffs.GetBuffSpec(buffId); b != nil {
			if b.Name == `empty` {
				continue
			}
			buffSpecs = append(buffSpecs, *b)
		}
	}
	sort.SliceStable(buffSpecs, func(i, j int) bool {
		return buffSpecs[i].BuffId < buffSpecs[j].BuffId
	})

	tplData := map[string]any{}

	tplData[`mobInfo`] = *mobInfo

	shopData := map[string]characters.Shop{
		`Items`:       {},
		`Buffs`:       {},
		`Mercenaries`: {},
		`Pets`:        {},
	}

	for _, shopItm := range mobInfo.Character.Shop {

		if shopItm.ItemId > 0 {
			shopData[`Items`] = append(shopData[`Items`], shopItm)
			continue
		}

		if shopItm.BuffId > 0 {
			shopData[`Buffs`] = append(shopData[`Buffs`], shopItm)
			continue
		}

		if shopItm.MobId > 0 {
			shopData[`Mercenaries`] = append(shopData[`Mercenaries`], shopItm)
			continue
		}

		if shopItm.PetType != `` {
			shopData[`Pets`] = append(shopData[`Pets`], shopItm)
			continue
		}
	}
	tplData[`mobShop`] = shopData

	tplData[`characterInfo`] = &mobInfo.Character
	tplData[`allZoneNames`] = allZoneNames
	tplData[`allRaces`] = allRaces
	tplData[`activityLevels`] = activityLevels
	tplData[`dropChances`] = dropChances
	tplData[`allMobGroups`] = allMobGroups
	tplData[`buffSpecs`] = buffSpecs

	if err := tmpl.Execute(w, tplData); err != nil {
		mudlog.Error("HTML Execute", "error", err)
	}

}
