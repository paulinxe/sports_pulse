// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {EIP712} from "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import {CompetitionRegistry} from "./CompetitionRegistry.sol";
import {TeamRegistry} from "./TeamRegistry.sol";

contract MatchRegistry is EIP712 {

    struct Match {
        uint64 matchId;
        uint32 homeTeamId;
        uint32 awayTeamId;
        uint8 homeTeamScore;
        uint8 awayTeamScore;
    }

    // The address of the verified signer who signs matches
    address public immutable verifiedSigner;

    // We need the address of each Registry so we can query to validate the data
    CompetitionRegistry public immutable competitionRegistry;
    TeamRegistry public immutable teamRegistry;

    // We don't allow scores higher than 80
    uint8 constant MAX_SCORE = 80;

    error InvalidTeams(uint32 homeTeamId, uint32 awayTeamId);
    error InvalidMatchId(bytes32 matchId);
    error InvalidCompetitionId(uint32 competitionId);
    error InvalidHomeTeamId(uint32 homeTeamId);
    error InvalidAwayTeamId(uint32 awayTeamId);
    error InvalidMatchDate(uint32 matchDate);
    error InvalidScores(uint8 homeTeamScore, uint8 awayTeamScore);

    constructor(
        address _verifiedSigner,
        CompetitionRegistry _competitionRegistry,
        TeamRegistry _teamRegistry
    ) EIP712("SportsPulse", "1") {
        verifiedSigner = _verifiedSigner;
        competitionRegistry = _competitionRegistry;
        teamRegistry = _teamRegistry;
    }

    // The matchDate must be formatted as YYYYMMDD UTC time
    function submitMatch(
        bytes calldata _matchId,
        uint32 competitionId,
        uint32 homeTeamId,
        uint32 awayTeamId,
        uint8 homeTeamScore,
        uint8 awayTeamScore,
        uint32 matchDate,
        bytes calldata signature
    ) external {
        if (homeTeamId == awayTeamId) {
            revert InvalidTeams(homeTeamId, awayTeamId);
        }

        bytes32 matchId = bytes32(_matchId);
        if (keccak256(abi.encodePacked(competitionId, homeTeamId, awayTeamId, matchDate)) != matchId) {
            revert InvalidMatchId(matchId);
        }

        if (matchDate < 20100101 || matchDate > 21001231) {
            // We do care more about the length of the int. Making sure we only accept YYYYMMDD format.
            revert InvalidMatchDate(matchDate);
        }

        if (bytes(competitionRegistry.competitions(competitionId)).length == 0) {
            revert InvalidCompetitionId(competitionId);
        }

        if (bytes(teamRegistry.teams(homeTeamId)).length == 0) {
            revert InvalidHomeTeamId(homeTeamId);
        }

        if (bytes(teamRegistry.teams(awayTeamId)).length == 0) {
            revert InvalidAwayTeamId(awayTeamId);
        }

        if (homeTeamScore > MAX_SCORE || awayTeamScore > MAX_SCORE) {
            revert InvalidScores(homeTeamScore, awayTeamScore);
        }
    }
}