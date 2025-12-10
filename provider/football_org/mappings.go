package football_org

import "provider/entity"

var CompetitionToFootballOrgID = map[entity.Competition]int{
	entity.LaLiga: 2014,
}

// TODO: check if we can make this a constant.
var FootballOrgTeamMapping = map[int]entity.Team{
    263: entity.Alaves,
    77: entity.AthleticClub,
    78: entity.AtleticoMadrid,
    81: entity.Barcelona,
    558: entity.CeltaVigo,
    285: entity.Elche,
    80: entity.Espanyol,
    82: entity.Getafe,
    298: entity.Girona,
    88: entity.Levante,
    89: entity.Mallorca,
    79: entity.Osasuna,
    1048: entity.Oviedo,
    87: entity.RayoVallecano,
    90: entity.RealBetis,
    86: entity.RealMadrid,
    92: entity.RealSociedad,
    559: entity.Sevilla,
    95: entity.Valencia,
    94: entity.Villarreal,
}