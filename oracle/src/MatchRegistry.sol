// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {EIP712} from "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {CompetitionRegistry} from "./CompetitionRegistry.sol";
import {TeamRegistry} from "./TeamRegistry.sol";
import {console} from "forge-std/console.sol";

contract MatchRegistry is EIP712 {
    using ECDSA for bytes32;

    // The match data.
    // We don't need the competitionId, homeTeamId, and awayTeamId because we can derive them from the matchId.
    struct Match {
        bytes32 matchId;
        uint8 homeTeamScore;
        uint8 awayTeamScore;
    }

    // The address of the verified signer who signs matches
    address public immutable authorizedSigner;
    // We need the address of each Registry so we can query to validate the data
    CompetitionRegistry public immutable competitionRegistry;
    TeamRegistry public immutable teamRegistry;
    // We don't allow scores higher than 80
    uint8 constant MAX_SCORE = 80;
    bytes32 public constant MATCH_RESULT_TYPEHASH = keccak256("Match(bytes32 matchId,uint8 homeScore,uint8 awayScore)");
    mapping(bytes32 => Match) public matches;

    event MatchRegistered(bytes32 indexed matchId, uint8 homeTeamScore, uint8 awayTeamScore);

    error InvalidTeams(uint32 homeTeamId, uint32 awayTeamId);
    error InvalidMatchId(bytes32 matchId);
    error InvalidCompetitionId(uint32 competitionId);
    error InvalidHomeTeamId(uint32 homeTeamId);
    error InvalidAwayTeamId(uint32 awayTeamId);
    error InvalidMatchDate(uint32 matchDate);
    error InvalidScores(uint8 homeTeamScore, uint8 awayTeamScore);
    error InvalidSignature(bytes signature);

    constructor(
        address _authorizedSigner,
        CompetitionRegistry _competitionRegistry,
        TeamRegistry _teamRegistry
    ) EIP712("SportsPulse", "1") {
        authorizedSigner = _authorizedSigner;
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

        validateSignature(matchId, homeTeamScore, awayTeamScore, signature);

        matches[matchId] = Match(matchId, homeTeamScore, awayTeamScore);

        emit MatchRegistered(matchId, homeTeamScore, awayTeamScore);
    }

    function validateSignature(bytes32 matchId, uint8 homeTeamScore, uint8 awayTeamScore, bytes calldata signature) private view {
        bytes32 structHash = keccak256(
            abi.encode(
                MATCH_RESULT_TYPEHASH,
                matchId,
                homeTeamScore,
                awayTeamScore
            )
        );
        
        console.log("=== Solidity EIP-712 Hash Computation ===");
        console.log("Struct hash:", uint256(structHash));
        console.log("MATCH_RESULT_TYPEHASH:", uint256(MATCH_RESULT_TYPEHASH));
        console.log("MatchId:", uint256(matchId));
        console.log("HomeScore:", homeTeamScore);
        console.log("AwayScore:", awayTeamScore);
        
        bytes32 domainSeparator = _domainSeparatorV4();
        console.log("Domain separator:", uint256(domainSeparator));
        
        bytes32 digest = _hashTypedDataV4(structHash);
        console.log("Final digest (toTypedDataHash):", uint256(digest));
        console.log("========================================");

        address signer = ECDSA.recoverCalldata(digest, signature);
        console.log("Recovered signer:", signer);
        console.log("Authorized signer:", authorizedSigner);
        if (signer != authorizedSigner) {
            revert InvalidSignature(signature);
        }
    }
}