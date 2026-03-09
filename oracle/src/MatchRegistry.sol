// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {CompetitionRegistry} from "./CompetitionRegistry.sol";
import {TeamRegistry} from "./TeamRegistry.sol";

contract MatchRegistry is Ownable {
    // slither-disable-next-line naming-convention
    CompetitionRegistry public immutable COMPETITION_REGISTRY;
    // slither-disable-next-line naming-convention
    TeamRegistry public immutable TEAM_REGISTRY;

    struct MatchInput {
        uint8 competitionId;
        uint16 seasonStartYear;
        uint8 journeyNumber;
        uint16 homeTeamId;
        uint16 awayTeamId;
    }

    uint8 public constant MAX_BATCH_SIZE = 50;
    mapping(bytes32 => bool) public registeredMatches;

    event MatchRegistered(bytes32 indexed matchId);

    error BatchTooLarge(uint256 length);
    error InvalidCompetitionId(uint8 competitionId);
    error InvalidHomeTeamId(uint16 homeTeamId);
    error InvalidAwayTeamId(uint16 awayTeamId);
    error InvalidTeams(uint16 homeTeamId, uint16 awayTeamId);
    error MatchAlreadyRegistered(bytes32 matchId);

    constructor(CompetitionRegistry _competitionRegistry, TeamRegistry _teamRegistry) Ownable(msg.sender) {
        COMPETITION_REGISTRY = _competitionRegistry;
        TEAM_REGISTRY = _teamRegistry;
    }

    /**
     * @notice Register a batch of matches. Reverts entirely if any match is invalid or duplicate.
     * @param matches Up to MAX_BATCH_SIZE matches to register.
     */
    function registerBatch(MatchInput[] calldata matches) external onlyOwner {
        if (matches.length > MAX_BATCH_SIZE) {
            revert BatchTooLarge(matches.length);
        }

        // slither-disable-start calls-loop
        // COMPETITION_REGISTRY and TEAM_REGISTRY are immutable contracts under our control; external calls in loop are intentional.
        for (uint256 i = 0; i < matches.length; i++) {
            MatchInput calldata m = matches[i];

            if (m.homeTeamId == m.awayTeamId) {
                revert InvalidTeams(m.homeTeamId, m.awayTeamId);
            }

            if (bytes(COMPETITION_REGISTRY.competitions(m.competitionId)).length == 0) {
                revert InvalidCompetitionId(m.competitionId);
            }

            if (bytes(TEAM_REGISTRY.teams(m.homeTeamId)).length == 0) {
                revert InvalidHomeTeamId(m.homeTeamId);
            }

            if (bytes(TEAM_REGISTRY.teams(m.awayTeamId)).length == 0) {
                revert InvalidAwayTeamId(m.awayTeamId);
            }

            // TODO: two teams play twice per season only. We need to check that the match is not already registered for other journeys of the same season.

            bytes32 matchId =
                _getMatchId(m.competitionId, m.seasonStartYear, m.journeyNumber, m.homeTeamId, m.awayTeamId);

            if (registeredMatches[matchId]) {
                revert MatchAlreadyRegistered(matchId);
            }

            registeredMatches[matchId] = true;

            emit MatchRegistered(matchId);
        }
        // slither-disable-end calls-loop
    }

    function _getMatchId(
        uint8 competitionId,
        uint16 seasonStartYear,
        uint8 journeyNumber,
        uint16 homeTeamId,
        uint16 awayTeamId
    ) private pure returns (bytes32) {
        return keccak256(abi.encodePacked(competitionId, seasonStartYear, journeyNumber, homeTeamId, awayTeamId));
    }
}
