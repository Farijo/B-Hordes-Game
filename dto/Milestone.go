package dto

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
		Wid        jsonNullByte `db:"mapWid"`
		Hei        jsonNullByte `db:"mapHei"`
		Days       jsonNullByte `db:"mapDays"`
		Conspiracy jsonNullBool `db:"conspiracy"`
		Custom     jsonNullBool `db:"custom"`
		City       struct {
			Buildings jsonNullList `db:"buildings"`
			Bank      jsonNullDict `db:"bank"`
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
	if incoming.Map.Custom.Valid {
		incoming.Map.Custom.Valid = !previous.Map.Custom.Valid || (incoming.Map.Custom.Bool != previous.Map.Custom.Bool)
		hasChanges = hasChanges || incoming.Map.Custom.Valid
	}

	hasChanges = incoming.Rewards.KeepDifferencesOnly(previous.Rewards) || hasChanges
	hasChanges = incoming.Map.City.Bank.KeepDifferencesOnly(previous.Map.City.Bank) || hasChanges
	hasChanges = incoming.Map.City.Buildings.KeepDifferencesOnly(previous.Map.City.Buildings) || hasChanges
	hasChanges = incoming.Map.Zones.KeepDifferencesOnly(previous.Map.Zones) || hasChanges

	return hasChanges
}
