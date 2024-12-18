package dto

import "strings"

type Milestone struct {
	User       User            `db:"user"`
	Dt         string          `db:"dt"`
	IsGhost    jsonNullBool    `db:"isGhost"`
	PlayedMaps jsonNullCounter `db:"playedMaps"`
	Rewards    jsonNullDict    `db:"rewards"`
	Dead       jsonNullBool    `db:"dead"`
	Out        jsonNullBool    `db:"isOut"`
	Ban        jsonNullBool    `db:"ban"`
	BaseDef    jsonNullByte    `db:"baseDef"`
	X          jsonNullInt16   `db:"x"`
	Y          jsonNullInt16   `db:"y"`
	Job        jsonNullJob     `db:"job"`
	Map        struct {
		Wid        jsonNullByte  `db:"mapWid"`
		Hei        jsonNullByte  `db:"mapHei"`
		Days       jsonNullByte  `db:"mapDays"`
		Conspiracy jsonNullBool  `db:"conspiracy"`
		Guide      jsonNullInt16 `db:"guide"`
		Shaman     jsonNullInt16 `db:"shaman"`
		Custom     jsonNullBool  `db:"custom"`
		City       struct {
			Door      jsonNullBool  `db:"door"`
			Water     jsonNullInt16 `db:"cityWater"`
			Chaos     jsonNullBool  `db:"chaos"`
			Devast    jsonNullBool  `db:"devast"`
			Hard      jsonNullBool  `db:"hard"`
			X         jsonNullInt16 `db:"cityX"`
			Y         jsonNullInt16 `db:"cityY"`
			Buildings jsonNullList  `db:"buildings"`
			News      jsonNullAux[struct {
				Z        jsonNullInt16 `db:"z"`
				Def      jsonNullInt16 `db:"def"`
				Water    jsonNullInt16 `db:"water"`
				RegenDir jsonNullRegen `db:"regenDir"`
			}]
			Defense struct {
				Total            jsonNullInt16 `db:"total"`
				Base             jsonNullInt16 `db:"base"`
				Buildings        jsonNullInt16 `db:"defBuildings"`
				Upgrades         jsonNullInt16 `db:"defUpgrades"`
				Items            jsonNullInt16 `db:"items"`
				ItemsMul         jsonNullFloat `db:"itemsMul"`
				CitizenHomes     jsonNullInt16 `db:"citizenHomes"`
				CitizenGuardians jsonNullInt16 `db:"citizenGuardians"`
				Watchmen         jsonNullInt16 `db:"watchmen"`
				Souls            jsonNullInt16 `db:"souls"`
				Temp             jsonNullInt16 `db:"temp"`
				Cadavers         jsonNullInt16 `db:"cadavers"`
				Bonus            jsonNullFloat `db:"bonus"`
			}
			Upgrades jsonNullAux[struct {
				List jsonNullDict `db:"upgrades"`
			}]
			Estimations,
			EstimationsNext jsonNullAux[struct {
				Min   jsonNullInt16 `db:"min"`
				Max   jsonNullInt16 `db:"max"`
				Maxed jsonNullBool  `db:"maxed"`
			}]
			Bank jsonNullDict `db:"bank"`
		}
		Zones jsonNullDict `db:"zoneItems"`
	}
}

func (incoming *Milestone) CheckFieldsDifference(previous *Milestone) bool {
	hasChanges := false

	if incoming.IsGhost.Valid {
		incoming.IsGhost.Valid = !previous.IsGhost.Valid || (incoming.IsGhost.Bool != previous.IsGhost.Bool)
		hasChanges = hasChanges || incoming.IsGhost.Valid
	}
	if incoming.PlayedMaps.Valid {
		incoming.PlayedMaps.Valid = !previous.PlayedMaps.Valid || (incoming.PlayedMaps.Int64 != previous.PlayedMaps.Int64)
		hasChanges = hasChanges || incoming.PlayedMaps.Valid
	}
	if incoming.Dead.Valid {
		incoming.Dead.Valid = !previous.Dead.Valid || (incoming.Dead.Bool != previous.Dead.Bool)
		hasChanges = hasChanges || incoming.Dead.Valid
	}
	if incoming.Out.Valid {
		incoming.Out.Valid = !previous.Out.Valid || (incoming.Out.Bool != previous.Out.Bool)
		hasChanges = hasChanges || incoming.Out.Valid
	}
	if incoming.Ban.Valid {
		incoming.Ban.Valid = !previous.Ban.Valid || (incoming.Ban.Bool != previous.Ban.Bool)
		hasChanges = hasChanges || incoming.Ban.Valid
	}
	if incoming.BaseDef.Valid {
		incoming.BaseDef.Valid = !previous.BaseDef.Valid || (incoming.BaseDef.Byte != previous.BaseDef.Byte)
		hasChanges = hasChanges || incoming.BaseDef.Valid
	}
	if incoming.X.Valid {
		incoming.X.Valid = !previous.X.Valid || (incoming.X.Int16 != previous.X.Int16)
		hasChanges = hasChanges || incoming.X.Valid
	}
	if incoming.Y.Valid {
		incoming.Y.Valid = !previous.Y.Valid || (incoming.Y.Int16 != previous.Y.Int16)
		hasChanges = hasChanges || incoming.Y.Valid
	}
	if incoming.Job.Valid {
		incoming.Job.Valid = !previous.Job.Valid || (incoming.Job.Byte != previous.Job.Byte)
		hasChanges = hasChanges || incoming.Job.Valid
	}
	if incoming.Map.Wid.Valid {
		incoming.Map.Wid.Valid = !previous.Map.Wid.Valid || (incoming.Map.Wid.Byte != previous.Map.Wid.Byte)
		hasChanges = hasChanges || incoming.Map.Wid.Valid
	}
	if incoming.Map.Hei.Valid {
		incoming.Map.Hei.Valid = !previous.Map.Hei.Valid || (incoming.Map.Hei.Byte != previous.Map.Hei.Byte)
		hasChanges = hasChanges || incoming.Map.Hei.Valid
	}
	if incoming.Map.Days.Valid {
		incoming.Map.Days.Valid = !previous.Map.Days.Valid || (incoming.Map.Days.Byte != previous.Map.Days.Byte)
		hasChanges = hasChanges || incoming.Map.Days.Valid
	}
	if incoming.Map.Conspiracy.Valid {
		incoming.Map.Conspiracy.Valid = !previous.Map.Conspiracy.Valid || (incoming.Map.Conspiracy.Bool != previous.Map.Conspiracy.Bool)
		hasChanges = hasChanges || incoming.Map.Conspiracy.Valid
	}
	if incoming.Map.Guide.Valid {
		incoming.Map.Guide.Valid = !previous.Map.Guide.Valid || (incoming.Map.Guide.Int16 != previous.Map.Guide.Int16)
		hasChanges = hasChanges || incoming.Map.Guide.Valid
	}
	if incoming.Map.Shaman.Valid {
		incoming.Map.Shaman.Valid = !previous.Map.Shaman.Valid || (incoming.Map.Shaman.Int16 != previous.Map.Shaman.Int16)
		hasChanges = hasChanges || incoming.Map.Shaman.Valid
	}
	if incoming.Map.Custom.Valid {
		incoming.Map.Custom.Valid = !previous.Map.Custom.Valid || (incoming.Map.Custom.Bool != previous.Map.Custom.Bool)
		hasChanges = hasChanges || incoming.Map.Custom.Valid
	}
	if incoming.Map.City.Door.Valid {
		incoming.Map.City.Door.Valid = !previous.Map.City.Door.Valid || (incoming.Map.City.Door.Bool != previous.Map.City.Door.Bool)
		hasChanges = hasChanges || incoming.Map.City.Door.Valid
	}
	if incoming.Map.City.Water.Valid {
		incoming.Map.City.Water.Valid = !previous.Map.City.Water.Valid || (incoming.Map.City.Water.Int16 != previous.Map.City.Water.Int16)
		hasChanges = hasChanges || incoming.Map.City.Water.Valid
	}
	if incoming.Map.City.Chaos.Valid {
		incoming.Map.City.Chaos.Valid = !previous.Map.City.Chaos.Valid || (incoming.Map.City.Chaos.Bool != previous.Map.City.Chaos.Bool)
		hasChanges = hasChanges || incoming.Map.City.Chaos.Valid
	}
	if incoming.Map.City.Devast.Valid {
		incoming.Map.City.Devast.Valid = !previous.Map.City.Devast.Valid || (incoming.Map.City.Devast.Bool != previous.Map.City.Devast.Bool)
		hasChanges = hasChanges || incoming.Map.City.Devast.Valid
	}
	if incoming.Map.City.Hard.Valid {
		incoming.Map.City.Hard.Valid = !previous.Map.City.Hard.Valid || (incoming.Map.City.Hard.Bool != previous.Map.City.Hard.Bool)
		hasChanges = hasChanges || incoming.Map.City.Hard.Valid
	}
	if incoming.Map.City.X.Valid {
		incoming.Map.City.X.Valid = !previous.Map.City.X.Valid || (incoming.Map.City.X.Int16 != previous.Map.City.X.Int16)
		hasChanges = hasChanges || incoming.Map.City.X.Valid
	}
	if incoming.Map.City.Y.Valid {
		incoming.Map.City.Y.Valid = !previous.Map.City.Y.Valid || (incoming.Map.City.Y.Int16 != previous.Map.City.Y.Int16)
		hasChanges = hasChanges || incoming.Map.City.Y.Valid
	}
	if incoming.Map.City.News.V.Z.Valid {
		incoming.Map.City.News.V.Z.Valid = !previous.Map.City.News.V.Z.Valid || (incoming.Map.City.News.V.Z.Int16 != previous.Map.City.News.V.Z.Int16)
		hasChanges = hasChanges || incoming.Map.City.News.V.Z.Valid
	}
	if incoming.Map.City.News.V.Def.Valid {
		incoming.Map.City.News.V.Def.Valid = !previous.Map.City.News.V.Def.Valid || (incoming.Map.City.News.V.Def.Int16 != previous.Map.City.News.V.Def.Int16)
		hasChanges = hasChanges || incoming.Map.City.News.V.Def.Valid
	}
	if incoming.Map.City.News.V.Water.Valid {
		incoming.Map.City.News.V.Water.Valid = !previous.Map.City.News.V.Water.Valid || (incoming.Map.City.News.V.Water.Int16 != previous.Map.City.News.V.Water.Int16)
		hasChanges = hasChanges || incoming.Map.City.News.V.Water.Valid
	}
	if incoming.Map.City.News.V.RegenDir.Valid {
		incoming.Map.City.News.V.RegenDir.Valid = !previous.Map.City.News.V.RegenDir.Valid || (incoming.Map.City.News.V.RegenDir.Byte != previous.Map.City.News.V.RegenDir.Byte)
		hasChanges = hasChanges || incoming.Map.City.News.V.RegenDir.Valid
	}
	if incoming.Map.City.Defense.Total.Valid {
		incoming.Map.City.Defense.Total.Valid = !previous.Map.City.Defense.Total.Valid || (incoming.Map.City.Defense.Total.Int16 != previous.Map.City.Defense.Total.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.Total.Valid
	}
	if incoming.Map.City.Defense.Base.Valid {
		incoming.Map.City.Defense.Base.Valid = !previous.Map.City.Defense.Base.Valid || (incoming.Map.City.Defense.Base.Int16 != previous.Map.City.Defense.Base.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.Base.Valid
	}
	if incoming.Map.City.Defense.Buildings.Valid {
		incoming.Map.City.Defense.Buildings.Valid = !previous.Map.City.Defense.Buildings.Valid || (incoming.Map.City.Defense.Buildings.Int16 != previous.Map.City.Defense.Buildings.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.Buildings.Valid
	}
	if incoming.Map.City.Defense.Upgrades.Valid {
		incoming.Map.City.Defense.Upgrades.Valid = !previous.Map.City.Defense.Upgrades.Valid || (incoming.Map.City.Defense.Upgrades.Int16 != previous.Map.City.Defense.Upgrades.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.Upgrades.Valid
	}
	if incoming.Map.City.Defense.Items.Valid {
		incoming.Map.City.Defense.Items.Valid = !previous.Map.City.Defense.Items.Valid || (incoming.Map.City.Defense.Items.Int16 != previous.Map.City.Defense.Items.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.Items.Valid
	}
	if incoming.Map.City.Defense.ItemsMul.Valid {
		incoming.Map.City.Defense.ItemsMul.Valid = !previous.Map.City.Defense.ItemsMul.Valid || (incoming.Map.City.Defense.ItemsMul.V != previous.Map.City.Defense.ItemsMul.V)
		hasChanges = hasChanges || incoming.Map.City.Defense.ItemsMul.Valid
	}
	if incoming.Map.City.Defense.CitizenHomes.Valid {
		incoming.Map.City.Defense.CitizenHomes.Valid = !previous.Map.City.Defense.CitizenHomes.Valid || (incoming.Map.City.Defense.CitizenHomes.Int16 != previous.Map.City.Defense.CitizenHomes.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.CitizenHomes.Valid
	}
	if incoming.Map.City.Defense.CitizenGuardians.Valid {
		incoming.Map.City.Defense.CitizenGuardians.Valid = !previous.Map.City.Defense.CitizenGuardians.Valid || (incoming.Map.City.Defense.CitizenGuardians.Int16 != previous.Map.City.Defense.CitizenGuardians.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.CitizenGuardians.Valid
	}
	if incoming.Map.City.Defense.Watchmen.Valid {
		incoming.Map.City.Defense.Watchmen.Valid = !previous.Map.City.Defense.Watchmen.Valid || (incoming.Map.City.Defense.Watchmen.Int16 != previous.Map.City.Defense.Watchmen.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.Watchmen.Valid
	}
	if incoming.Map.City.Defense.Souls.Valid {
		incoming.Map.City.Defense.Souls.Valid = !previous.Map.City.Defense.Souls.Valid || (incoming.Map.City.Defense.Souls.Int16 != previous.Map.City.Defense.Souls.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.Souls.Valid
	}
	if incoming.Map.City.Defense.Temp.Valid {
		incoming.Map.City.Defense.Temp.Valid = !previous.Map.City.Defense.Temp.Valid || (incoming.Map.City.Defense.Temp.Int16 != previous.Map.City.Defense.Temp.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.Temp.Valid
	}
	if incoming.Map.City.Defense.Cadavers.Valid {
		incoming.Map.City.Defense.Cadavers.Valid = !previous.Map.City.Defense.Cadavers.Valid || (incoming.Map.City.Defense.Cadavers.Int16 != previous.Map.City.Defense.Cadavers.Int16)
		hasChanges = hasChanges || incoming.Map.City.Defense.Cadavers.Valid
	}
	if incoming.Map.City.Defense.Bonus.Valid {
		incoming.Map.City.Defense.Bonus.Valid = !previous.Map.City.Defense.Bonus.Valid || (incoming.Map.City.Defense.Bonus.V != previous.Map.City.Defense.Bonus.V)
		hasChanges = hasChanges || incoming.Map.City.Defense.Bonus.Valid
	}
	if incoming.Map.City.Estimations.V.Min.Valid {
		incoming.Map.City.Estimations.V.Min.Valid = !previous.Map.City.Estimations.V.Min.Valid || (incoming.Map.City.Estimations.V.Min.Int16 != previous.Map.City.Estimations.V.Min.Int16)
		hasChanges = hasChanges || incoming.Map.City.Estimations.V.Min.Valid
	}
	if incoming.Map.City.Estimations.V.Max.Valid {
		incoming.Map.City.Estimations.V.Max.Valid = !previous.Map.City.Estimations.V.Max.Valid || (incoming.Map.City.Estimations.V.Max.Int16 != previous.Map.City.Estimations.V.Max.Int16)
		hasChanges = hasChanges || incoming.Map.City.Estimations.V.Max.Valid
	}
	if incoming.Map.City.Estimations.V.Maxed.Valid {
		incoming.Map.City.Estimations.V.Maxed.Valid = !previous.Map.City.Estimations.V.Maxed.Valid || (incoming.Map.City.Estimations.V.Maxed.Bool != previous.Map.City.Estimations.V.Maxed.Bool)
		hasChanges = hasChanges || incoming.Map.City.Estimations.V.Maxed.Valid
	}
	if incoming.Map.City.EstimationsNext.V.Min.Valid {
		incoming.Map.City.EstimationsNext.V.Min.Valid = !previous.Map.City.EstimationsNext.V.Min.Valid || (incoming.Map.City.EstimationsNext.V.Min.Int16 != previous.Map.City.EstimationsNext.V.Min.Int16)
		hasChanges = hasChanges || incoming.Map.City.EstimationsNext.V.Min.Valid
	}
	if incoming.Map.City.EstimationsNext.V.Max.Valid {
		incoming.Map.City.EstimationsNext.V.Max.Valid = !previous.Map.City.EstimationsNext.V.Max.Valid || (incoming.Map.City.EstimationsNext.V.Max.Int16 != previous.Map.City.EstimationsNext.V.Max.Int16)
		hasChanges = hasChanges || incoming.Map.City.EstimationsNext.V.Max.Valid
	}
	if incoming.Map.City.EstimationsNext.V.Maxed.Valid {
		incoming.Map.City.EstimationsNext.V.Maxed.Valid = !previous.Map.City.EstimationsNext.V.Maxed.Valid || (incoming.Map.City.EstimationsNext.V.Maxed.Bool != previous.Map.City.EstimationsNext.V.Maxed.Bool)
		hasChanges = hasChanges || incoming.Map.City.EstimationsNext.V.Maxed.Valid
	}

	hasChanges = incoming.Rewards.KeepDifferencesOnly(previous.Rewards) || hasChanges
	hasChanges = incoming.Map.City.Buildings.KeepDifferencesOnly(previous.Map.City.Buildings) || hasChanges
	hasChanges = incoming.Map.City.Upgrades.V.List.KeepDifferencesOnly(previous.Map.City.Upgrades.V.List) || hasChanges
	hasChanges = incoming.Map.City.Bank.KeepDifferencesOnly(previous.Map.City.Bank) || hasChanges
	hasChanges = incoming.Map.Zones.KeepDifferencesOnly(previous.Map.Zones) || hasChanges

	return hasChanges
}

func (incoming *Milestone) HasData() bool {
	return incoming.IsGhost.Valid ||
		incoming.PlayedMaps.Valid ||
		incoming.Rewards.Valid ||
		incoming.Dead.Valid ||
		incoming.Out.Valid ||
		incoming.Ban.Valid ||
		incoming.BaseDef.Valid ||
		incoming.X.Valid ||
		incoming.Y.Valid ||
		incoming.Job.Valid ||
		incoming.Map.Wid.Valid ||
		incoming.Map.Hei.Valid ||
		incoming.Map.Days.Valid ||
		incoming.Map.Conspiracy.Valid ||
		incoming.Map.Guide.Valid ||
		incoming.Map.Shaman.Valid ||
		incoming.Map.Custom.Valid ||
		incoming.Map.City.Door.Valid ||
		incoming.Map.City.Water.Valid ||
		incoming.Map.City.Chaos.Valid ||
		incoming.Map.City.Devast.Valid ||
		incoming.Map.City.Hard.Valid ||
		incoming.Map.City.X.Valid ||
		incoming.Map.City.Y.Valid ||
		incoming.Map.City.News.V.Z.Valid ||
		incoming.Map.City.News.V.Def.Valid ||
		incoming.Map.City.News.V.Water.Valid ||
		incoming.Map.City.News.V.RegenDir.Valid ||
		incoming.Map.City.Defense.Total.Valid ||
		incoming.Map.City.Defense.Base.Valid ||
		incoming.Map.City.Defense.Buildings.Valid ||
		incoming.Map.City.Defense.Upgrades.Valid ||
		incoming.Map.City.Defense.Items.Valid ||
		incoming.Map.City.Defense.ItemsMul.Valid ||
		incoming.Map.City.Defense.CitizenHomes.Valid ||
		incoming.Map.City.Defense.CitizenGuardians.Valid ||
		incoming.Map.City.Defense.Watchmen.Valid ||
		incoming.Map.City.Defense.Souls.Valid ||
		incoming.Map.City.Defense.Temp.Valid ||
		incoming.Map.City.Defense.Cadavers.Valid ||
		incoming.Map.City.Defense.Bonus.Valid ||
		incoming.Map.City.Upgrades.V.List.Valid ||
		incoming.Map.City.Estimations.V.Min.Valid ||
		incoming.Map.City.Estimations.V.Max.Valid ||
		incoming.Map.City.Estimations.V.Maxed.Valid ||
		incoming.Map.City.EstimationsNext.V.Min.Valid ||
		incoming.Map.City.EstimationsNext.V.Max.Valid ||
		incoming.Map.City.EstimationsNext.V.Maxed.Valid ||
		incoming.Map.City.Buildings.Valid ||
		incoming.Map.City.Bank.Valid ||
		incoming.Map.Zones.Valid
}

var fields = []string{"isGhost", "playedMaps", "rewards", "dead", "isOut", "ban", "baseDef", "x", "y", "job", "mapWid", "mapHei", "mapDays", "conspiracy", "guide", "shaman", "custom", "door", "cityWater", "chaos", "devast", "hard", "cityX", "cityY", "buildings", "z", "def", "water", "regenDir", "total", "base", "defBuildings", "defUpgrades", "items", "itemsMul", "citizenHomes", "citizenGuardians", "watchmen", "souls", "temp", "cadavers", "bonus", "upgrades", "estiMin", "estiMax", "estiMaxed", "nextMin", "nextMax", "nextMaxed", "bank", "zoneItems"}

func MilestoneFields(prefix, suffix string) string {
	return prefix + strings.Join(fields, suffix+","+prefix) + suffix
}

func (milestone *Milestone) MilestoneScan() []any {
	return []any{
		&milestone.IsGhost,
		&milestone.PlayedMaps,
		&milestone.Rewards,
		&milestone.Dead,
		&milestone.Out,
		&milestone.Ban,
		&milestone.BaseDef,
		&milestone.X,
		&milestone.Y,
		&milestone.Job,
		&milestone.Map.Wid,
		&milestone.Map.Hei,
		&milestone.Map.Days,
		&milestone.Map.Conspiracy,
		&milestone.Map.Guide,
		&milestone.Map.Shaman,
		&milestone.Map.Custom,
		&milestone.Map.City.Door,
		&milestone.Map.City.Water,
		&milestone.Map.City.Chaos,
		&milestone.Map.City.Devast,
		&milestone.Map.City.Hard,
		&milestone.Map.City.X,
		&milestone.Map.City.Y,
		&milestone.Map.City.Buildings,
		&milestone.Map.City.News.V.Z,
		&milestone.Map.City.News.V.Def,
		&milestone.Map.City.News.V.Water,
		&milestone.Map.City.News.V.RegenDir,
		&milestone.Map.City.Defense.Total,
		&milestone.Map.City.Defense.Base,
		&milestone.Map.City.Defense.Buildings,
		&milestone.Map.City.Defense.Upgrades,
		&milestone.Map.City.Defense.Items,
		&milestone.Map.City.Defense.ItemsMul,
		&milestone.Map.City.Defense.CitizenHomes,
		&milestone.Map.City.Defense.CitizenGuardians,
		&milestone.Map.City.Defense.Watchmen,
		&milestone.Map.City.Defense.Souls,
		&milestone.Map.City.Defense.Temp,
		&milestone.Map.City.Defense.Cadavers,
		&milestone.Map.City.Defense.Bonus,
		&milestone.Map.City.Upgrades.V.List,
		&milestone.Map.City.Estimations.V.Min,
		&milestone.Map.City.Estimations.V.Max,
		&milestone.Map.City.Estimations.V.Maxed,
		&milestone.Map.City.EstimationsNext.V.Min,
		&milestone.Map.City.EstimationsNext.V.Max,
		&milestone.Map.City.EstimationsNext.V.Maxed,
		&milestone.Map.City.Bank,
		&milestone.Map.Zones}
}

func (milestone *Milestone) MilestoneExec() []any {
	return []any{
		milestone.IsGhost,
		milestone.PlayedMaps,
		milestone.Rewards,
		milestone.Dead,
		milestone.Out,
		milestone.Ban,
		milestone.BaseDef,
		milestone.X,
		milestone.Y,
		milestone.Job,
		milestone.Map.Wid,
		milestone.Map.Hei,
		milestone.Map.Days,
		milestone.Map.Conspiracy,
		milestone.Map.Guide,
		milestone.Map.Shaman,
		milestone.Map.Custom,
		milestone.Map.City.Door,
		milestone.Map.City.Water,
		milestone.Map.City.Chaos,
		milestone.Map.City.Devast,
		milestone.Map.City.Hard,
		milestone.Map.City.X,
		milestone.Map.City.Y,
		milestone.Map.City.Buildings,
		milestone.Map.City.News.V.Z,
		milestone.Map.City.News.V.Def,
		milestone.Map.City.News.V.Water,
		milestone.Map.City.News.V.RegenDir,
		milestone.Map.City.Defense.Total,
		milestone.Map.City.Defense.Base,
		milestone.Map.City.Defense.Buildings,
		milestone.Map.City.Defense.Upgrades,
		milestone.Map.City.Defense.Items,
		milestone.Map.City.Defense.ItemsMul,
		milestone.Map.City.Defense.CitizenHomes,
		milestone.Map.City.Defense.CitizenGuardians,
		milestone.Map.City.Defense.Watchmen,
		milestone.Map.City.Defense.Souls,
		milestone.Map.City.Defense.Temp,
		milestone.Map.City.Defense.Cadavers,
		milestone.Map.City.Defense.Bonus,
		milestone.Map.City.Upgrades.V.List,
		milestone.Map.City.Estimations.V.Min,
		milestone.Map.City.Estimations.V.Max,
		milestone.Map.City.Estimations.V.Maxed,
		milestone.Map.City.EstimationsNext.V.Min,
		milestone.Map.City.EstimationsNext.V.Max,
		milestone.Map.City.EstimationsNext.V.Maxed,
		milestone.Map.City.Bank,
		milestone.Map.Zones}
}
