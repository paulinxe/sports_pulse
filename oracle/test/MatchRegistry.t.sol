// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";
import {MatchRegistry} from "../src/MatchRegistry.sol";
import {CompetitionRegistry} from "../src/CompetitionRegistry.sol";
import {TeamRegistry} from "../src/TeamRegistry.sol";

contract MatchRegistryTest is Test {
    MatchRegistry public matchRegistry;
    CompetitionRegistry public competitionRegistry;
    TeamRegistry public teamRegistry;

    uint8 constant COMPETITION_ID = 1;
    uint16 constant HOME_TEAM_ID = 1;
    uint16 constant AWAY_TEAM_ID = 2;
    uint16 constant SEASON_YEAR = 2025;
    uint8 constant JOURNEY = 1;

    function setUp() public {
        string[] memory competitionNames = new string[](1);
        competitionNames[0] = "Liga AUF Uruguay";
        competitionRegistry = new CompetitionRegistry(competitionNames);

        string[] memory teamNames = new string[](2);
        teamNames[0] = "Nacional";
        teamNames[1] = "Rampla Juniors";
        teamRegistry = new TeamRegistry(teamNames);

        matchRegistry = new MatchRegistry(competitionRegistry, teamRegistry);
    }

    function test_registerBatch_reverts_when_batch_too_big() public {
        uint256 tooMany = matchRegistry.MAX_BATCH_SIZE() + 1;
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](tooMany);
        for (uint256 i = 0; i < tooMany; i++) {
            inputs[i] = MatchRegistry.MatchInput({
                competitionId: COMPETITION_ID,
                seasonStartYear: SEASON_YEAR,
                journeyNumber: uint8(i + 1),
                homeTeamId: HOME_TEAM_ID,
                awayTeamId: AWAY_TEAM_ID
            });
        }

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.BatchTooLarge.selector, tooMany));
        matchRegistry.registerBatch(inputs);
    }

    function test_registerBatch_reverts_when_same_home_and_away_team() public {
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](1);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: HOME_TEAM_ID
        });

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidTeams.selector, HOME_TEAM_ID, HOME_TEAM_ID));
        matchRegistry.registerBatch(inputs);
    }

    function test_registerBatch_reverts_when_invalid_competition() public {
        uint8 invalidCompetitionId = 200;
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](1);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: invalidCompetitionId,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidCompetitionId.selector, invalidCompetitionId));
        matchRegistry.registerBatch(inputs);
    }

    function test_registerBatch_reverts_when_invalid_home_team() public {
        uint16 invalidHomeTeamId = 999;
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](1);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: invalidHomeTeamId,
            awayTeamId: AWAY_TEAM_ID
        });

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidHomeTeamId.selector, invalidHomeTeamId));
        matchRegistry.registerBatch(inputs);
    }

    function test_registerBatch_reverts_when_invalid_away_team() public {
        uint16 invalidAwayTeamId = 999;
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](1);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: invalidAwayTeamId
        });

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidAwayTeamId.selector, invalidAwayTeamId));
        matchRegistry.registerBatch(inputs);
    }

    function test_registerBatch_reverts_when_match_already_registered() public {
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](1);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });

        matchRegistry.registerBatch(inputs);

        bytes32 expectedMatchId = _computeMatchId(COMPETITION_ID, SEASON_YEAR, JOURNEY, HOME_TEAM_ID, AWAY_TEAM_ID);
        vm.expectRevert(
            abi.encodeWithSelector(MatchRegistry.MatchAlreadyRegistered.selector, expectedMatchId)
        );
        matchRegistry.registerBatch(inputs);
    }

    function test_registerBatch_reverts_entire_batch_when_second_invalid() public {
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](2);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: 1,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });
        inputs[1] = MatchRegistry.MatchInput({
            competitionId: 200,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: 2,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });

        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.InvalidCompetitionId.selector, uint8(200)));
        matchRegistry.registerBatch(inputs);

        bytes32 firstMatchId = _computeMatchId(COMPETITION_ID, SEASON_YEAR, 1, HOME_TEAM_ID, AWAY_TEAM_ID);
        assertFalse(matchRegistry.registeredMatches(firstMatchId));
    }

    function test_registerBatch_reverts_when_duplicate_in_same_batch() public {
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](2);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });
        inputs[1] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });

        bytes32 expectedMatchId = _computeMatchId(COMPETITION_ID, SEASON_YEAR, JOURNEY, HOME_TEAM_ID, AWAY_TEAM_ID);
        vm.expectRevert(abi.encodeWithSelector(MatchRegistry.MatchAlreadyRegistered.selector, expectedMatchId));
        matchRegistry.registerBatch(inputs);
    }

    function test_only_owner_can_registerBatch() public {
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](1);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });

        vm.prank(makeAddr("stranger"));
        vm.expectRevert();
        matchRegistry.registerBatch(inputs);
    }

    function test_registerBatch_single_match_success() public {
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](1);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });

        bytes32 expectedMatchId = _computeMatchId(COMPETITION_ID, SEASON_YEAR, JOURNEY, HOME_TEAM_ID, AWAY_TEAM_ID);

        vm.expectEmit(true, true, true, true);
        emit MatchRegistry.MatchRegistered(expectedMatchId);

        matchRegistry.registerBatch(inputs);

        assertTrue(matchRegistry.registeredMatches(expectedMatchId));
    }

    function test_registerBatch_emits_MatchRegistered_per_match() public {
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](3);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: 1,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });
        inputs[1] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: 2,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });
        inputs[2] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: 3,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });

        for (uint256 i = 0; i < 3; i++) {
            vm.expectEmit(true, true, true, true);
            emit MatchRegistry.MatchRegistered(
                _computeMatchId(
                    inputs[i].competitionId,
                    inputs[i].seasonStartYear,
                    inputs[i].journeyNumber,
                    inputs[i].homeTeamId,
                    inputs[i].awayTeamId
                )
            );
        }
        matchRegistry.registerBatch(inputs);

        for (uint256 i = 0; i < 3; i++) {
            assertTrue(
                matchRegistry.registeredMatches(
                    _computeMatchId(
                        inputs[i].competitionId,
                        inputs[i].seasonStartYear,
                        inputs[i].journeyNumber,
                        inputs[i].homeTeamId,
                        inputs[i].awayTeamId
                    )
                )
            );
        }
    }

    function test_registerBatch_max_50_succeeds() public {
        uint256 n = matchRegistry.MAX_BATCH_SIZE();
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](n);
        for (uint256 i = 0; i < n; i++) {
            inputs[i] = MatchRegistry.MatchInput({
                competitionId: COMPETITION_ID,
                seasonStartYear: SEASON_YEAR,
                journeyNumber: uint8(i + 1),
                homeTeamId: HOME_TEAM_ID,
                awayTeamId: AWAY_TEAM_ID
            });
        }

        matchRegistry.registerBatch(inputs);

        for (uint256 i = 0; i < n; i++) {
            assertTrue(
                matchRegistry.registeredMatches(
                    _computeMatchId(
                        COMPETITION_ID,
                        SEASON_YEAR,
                        uint8(i + 1),
                        HOME_TEAM_ID,
                        AWAY_TEAM_ID
                    )
                )
            );
        }
    }

    /// Same formula as MatchRegistry: keccak256(abi.encodePacked(competitionId, seasonStartYear, journeyNumber, homeTeamId, awayTeamId))
    function _computeMatchId(
        uint8 competitionId,
        uint16 seasonStartYear,
        uint8 journeyNumber,
        uint16 homeTeamId,
        uint16 awayTeamId
    ) private pure returns (bytes32) {
        return keccak256(
            abi.encodePacked(competitionId, seasonStartYear, journeyNumber, homeTeamId, awayTeamId)
        );
    }
}
