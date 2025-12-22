// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {Test} from "forge-std/Test.sol";
import {MatchRegistry} from "../src/MatchRegistry.sol";
import {CompetitionRegistry} from "../src/CompetitionRegistry.sol";
import {TeamRegistry} from "../src/TeamRegistry.sol";
import {console} from "forge-std/console.sol";

contract MatchRegistryTest is Test {
    MatchRegistry public matchRegistry;

    uint32 constant COMPETITION_ID = 1;
    uint32 constant HOME_TEAM_ID = 1;
    uint32 constant AWAY_TEAM_ID = 2;

    function setUp() public {
        string[] memory competitionNames = new string[](1);
        competitionNames[COMPETITION_ID - 1] = "LaLiga";
        CompetitionRegistry competitionRegistry = new CompetitionRegistry(competitionNames);

        string[] memory teamNames = new string[](2);
        teamNames[HOME_TEAM_ID - 1] = "Nacional";
        teamNames[AWAY_TEAM_ID - 1] = "Basanez";
        TeamRegistry teamRegistry = new TeamRegistry(teamNames);

        address verifiedSigner = 0xCC0724CDc18DaE6B469b8e8B533fCd4dBE32FB46;
        matchRegistry = new MatchRegistry(verifiedSigner, competitionRegistry, teamRegistry);
    }

    function test_submit_reverts_when_invalid_teams() public {
        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidTeams.selector, HOME_TEAM_ID, HOME_TEAM_ID));
        matchRegistry.submitMatch(abi.encodePacked(""), COMPETITION_ID, HOME_TEAM_ID, HOME_TEAM_ID, 1, 1, 1, "");
    }

    function test_submit_reverts_when_invalid_match_id() public {
        bytes memory matchId = abi.encodePacked("invalid_match_id");
        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidMatchId.selector, bytes32(matchId)));
        matchRegistry.submitMatch(matchId, COMPETITION_ID, HOME_TEAM_ID, AWAY_TEAM_ID, 1, 1, 1, "");
    }

    function testFuzz_submit_reverts_when_invalid_match_date(uint32 matchDate) public {
        vm.assume(matchDate < 20100101 || matchDate > 21001231);
        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidMatchDate.selector, matchDate));
        matchRegistry.submitMatch(generateMatchId(matchDate), COMPETITION_ID, HOME_TEAM_ID, AWAY_TEAM_ID, 1, 1, matchDate, "");
    }

    function test_submit_reverts_when_competition_does_not_exist() public {
        uint32 invalidCompetitionId = 999;
        uint32 matchDate = 20250101;

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidCompetitionId.selector, invalidCompetitionId));

        matchRegistry.submitMatch(generateMatchId(matchDate, invalidCompetitionId), invalidCompetitionId, HOME_TEAM_ID, AWAY_TEAM_ID, 1, 1, matchDate, "");
    }

    function test_submit_reverts_when_home_team_does_not_exist() public {
        uint32 invalidHomeTeamId = 999;
        uint32 matchDate = 20250101;

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidHomeTeamId.selector, invalidHomeTeamId));

        matchRegistry.submitMatch(generateMatchId(matchDate, invalidHomeTeamId, AWAY_TEAM_ID), COMPETITION_ID, invalidHomeTeamId, AWAY_TEAM_ID, 1, 1, matchDate, "");
    }

    function test_submit_reverts_when_away_team_does_not_exist() public {
        uint32 invalidAwayTeamId = 999;
        uint32 matchDate = 20250101;

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidAwayTeamId.selector, invalidAwayTeamId));

        matchRegistry.submitMatch(generateMatchId(matchDate, HOME_TEAM_ID, invalidAwayTeamId), COMPETITION_ID, HOME_TEAM_ID, invalidAwayTeamId, 1, 1, matchDate, "");
    }

    function testFuzz_submit_reverts_when_scores_are_higher_than_80(uint8 homeTeamScore, uint8 awayTeamScore) public {
        vm.assume(homeTeamScore > 80 || awayTeamScore > 80);
        uint32 matchDate = 20250101;

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidScores.selector, homeTeamScore, awayTeamScore));

        matchRegistry.submitMatch(generateMatchId(matchDate), COMPETITION_ID, HOME_TEAM_ID, AWAY_TEAM_ID, homeTeamScore, awayTeamScore, matchDate, "");
    }

    function test_submit_reverts_when_signature_is_invalid() public {
        bytes memory signature = hex"c3dc2b81e3d1f01eb29edd0684cdf9acbd0fa0486dbb11621659507d8d4e5b9c59f3ff5d9b753a776802cde1bfd5a9d041df82e93a9f7efa3880d9015c44552801";
        uint32 matchDate = 20250101;

        vm.expectRevert(abi.encodeWithSelector(ECDSA.ECDSAInvalidSignature.selector));

        matchRegistry.submitMatch(generateMatchId(matchDate), COMPETITION_ID, HOME_TEAM_ID, AWAY_TEAM_ID, 1, 1, matchDate, signature);
    }

    function test_submit_reverts_when_signature_format_is_valid_but_signer_is_not_allowed() public {
        bytes memory signature = hex"cea04c025cd2c580f76d1eb92199e900ee82fe12d3a23979e25d455bfbeb275f5ca198283f8bda7f6679c2a3847d6da34c47303f2830a08d6016199b9584478e1c";
        uint32 matchDate = 20251219;
        uint8 homeTeamScore = 1;
        uint8 awayTeamScore = 2;

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidSignature.selector, signature));

        matchRegistry.submitMatch(generateMatchId(matchDate), COMPETITION_ID, HOME_TEAM_ID, AWAY_TEAM_ID, homeTeamScore, awayTeamScore, matchDate, signature);
    }

    function test_we_store_the_match() public {
        // The given signature comes from the Signer go test: Test_we_sign_a_match
        bytes memory signature = hex"4f0fa54d6dd9629d5f1d6b0f17236f4f9f009b72be6e77bdc56a4d0d891c0c076f6c36472f7b667d5f63895424a19a19bc56f264e49699c58bb07ec0868440081c";
        uint32 matchDate = 20251219;
        bytes memory matchId = generateMatchId(matchDate);
        uint8 homeTeamScore = 1;
        uint8 awayTeamScore = 2;
        
        vm.expectEmit(true, true, true, true);
        emit MatchRegistry.MatchRegistered(bytes32(matchId), homeTeamScore, awayTeamScore);
        
        matchRegistry.submitMatch(matchId, COMPETITION_ID, HOME_TEAM_ID, AWAY_TEAM_ID, homeTeamScore, awayTeamScore, matchDate, signature);

        (bytes32 storedMatchId, uint8 storedHomeTeamScore, uint8 storedAwayTeamScore) = matchRegistry.matches(bytes32(matchId));
        assertEq(storedMatchId, bytes32(matchId));
        assertEq(storedHomeTeamScore, homeTeamScore);
        assertEq(storedAwayTeamScore, awayTeamScore);
    }

    function generateMatchId(uint32 matchDate) private pure returns (bytes memory) {
        return abi.encodePacked(keccak256(abi.encodePacked(COMPETITION_ID, HOME_TEAM_ID, AWAY_TEAM_ID, matchDate)));
    }

    function generateMatchId(uint32 matchDate, uint32 competitionId) private pure returns (bytes memory) {
        return abi.encodePacked(keccak256(abi.encodePacked(competitionId, HOME_TEAM_ID, AWAY_TEAM_ID, matchDate)));
    }

    function generateMatchId(uint32 matchDate, uint32 homeTeamId, uint32 awayTeamId) private pure returns (bytes memory) {
        return abi.encodePacked(keccak256(abi.encodePacked(COMPETITION_ID, homeTeamId, awayTeamId, matchDate)));
    }
}