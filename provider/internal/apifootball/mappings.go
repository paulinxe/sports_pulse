package apifootball

import "provider/internal/entity"

const LeagueIDChampionship = "153"

var CompetitionToAPIFootballID = map[entity.Competition]string{
	entity.Championship: LeagueIDChampionship,
}

var APIFootballIDToCompetition = map[string]entity.Competition{
	LeagueIDChampionship: entity.Championship,
}

var APIFootballTeamMapping = map[string]entity.Team{
	"3432": entity.Birmingham,
	"3096": entity.Blackburn,
	"3084": entity.BristolCity,
	"3117": entity.Charlton,
	"3094": entity.Coventry,
	"3426": entity.Derby,
	"3119": entity.Hull,
	"3121": entity.Ipswich,
	"155": entity.Leicester,
	"3425": entity.Middlesbrough,
	"3083": entity.Millwall,
	"3093": entity.Norwich,
	"3113": entity.OxfordUtd,
	"3105": entity.Portsmouth,
	"3424": entity.Preston,
	"3098": entity.QPR,
	"3074": entity.SheffieldUtd,
	"3099": entity.SheffieldWed,
	"3072": entity.Southampton,
	"3097": entity.Stoke,
	"3076": entity.Swansea,
	"3427": entity.Watford,
	"3423": entity.WestBrom,
	"2954": entity.Wrexham,
}
