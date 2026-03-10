// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {ResultRegistry} from "../src/ResultRegistry.sol";
import {MatchRegistry} from "../src/MatchRegistry.sol";
import {CompetitionRegistry} from "../src/CompetitionRegistry.sol";
import {TeamRegistry} from "../src/TeamRegistry.sol";

contract ResultRegistryTest is Test {
    ResultRegistry public resultRegistry;
    MatchRegistry public matchRegistry;
    CompetitionRegistry public competitionRegistry;
    TeamRegistry public teamRegistry;

    address public owner;
    address public nonOwner;

    uint8 constant COMPETITION_ID = 1;
    uint16 constant HOME_TEAM_ID = 1;
    uint16 constant AWAY_TEAM_ID = 2;
    uint16 constant SEASON_YEAR = 2025;
    uint8 constant JOURNEY = 1;

    bytes32 matchId;

    function setUp() public {
        owner = makeAddr("owner");
        nonOwner = makeAddr("nonOwner");

        string[] memory competitionNames = new string[](1);
        competitionNames[0] = "Liga AUF Uruguay";
        competitionRegistry = new CompetitionRegistry(competitionNames);

        string[] memory teamNames = new string[](2);
        teamNames[0] = "Nacional";
        teamNames[1] = "Rampla Juniors";
        teamRegistry = new TeamRegistry(teamNames);

        matchRegistry = new MatchRegistry(competitionRegistry, teamRegistry);

        // Setup ResultRegistry with owner
        vm.prank(owner);
        resultRegistry = new ResultRegistry(address(matchRegistry));

        // Compute a valid matchId for tests
        matchId = keccak256(abi.encodePacked(COMPETITION_ID, SEASON_YEAR, JOURNEY, HOME_TEAM_ID, AWAY_TEAM_ID));
    }

    // --- signalMatchEnd() ---

    function test_signalMatchEnd_reverts_when_caller_not_owner() public {
        _registerMatch();

        vm.prank(nonOwner);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, nonOwner));
        resultRegistry.signalMatchEnd(matchId);
    }

    function test_signalMatchEnd_reverts_when_match_not_registered() public {
        bytes32 unregisteredMatchId = keccak256("unregistered");

        vm.prank(owner);
        vm.expectRevert(abi.encodeWithSelector(ResultRegistry.MatchNotRegistered.selector, unregisteredMatchId));
        resultRegistry.signalMatchEnd(unregisteredMatchId);
    }

    function test_signalMatchEnd_reverts_when_already_opened() public {
        _registerMatch();

        // First call succeeds
        vm.prank(owner);
        resultRegistry.signalMatchEnd(matchId);

        // Second call reverts
        vm.prank(owner);
        vm.expectRevert(abi.encodeWithSelector(ResultRegistry.SubmissionWindowAlreadyOpened.selector, matchId));
        resultRegistry.signalMatchEnd(matchId);
    }

    /**
     * @dev Fuzz test for all non-INACTIVE statuses that should prevent re-signalling.
     * Statuses: OPEN (1), FINALISED (2), DISPUTED (3), UNRESOLVABLE (4)
     */
    function testFuzz_signalMatchEnd_fails_if_status_not_inactive(uint8 status) public {
        // Bound status to valid non-INACTIVE statuses (1-4)
        status = uint8(bound(status, 1, 4));

        _registerMatch();

        // First open the window
        vm.prank(owner);
        resultRegistry.signalMatchEnd(matchId);

        // Manually manipulate storage to set status to the test parameter
        vm.store(
            address(resultRegistry),
            keccak256(abi.encodePacked(matchId, uint256(0))), // slot for results[matchId].status
            bytes32(uint256(status))
        );

        // Try to signal again - should revert because status is not INACTIVE
        vm.prank(owner);
        vm.expectRevert(abi.encodeWithSelector(ResultRegistry.SubmissionWindowAlreadyOpened.selector, matchId));
        resultRegistry.signalMatchEnd(matchId);
    }

    function test_signalMatchEnd_sets_correct_values() public {
        _registerMatch();

        uint256 beforeTimestamp = block.timestamp;
        uint256 expectedWindowClosesAt = beforeTimestamp + resultRegistry.SUBMISSION_WINDOW();

        vm.prank(owner);
        vm.expectEmit(true, true, true, true);
        emit ResultRegistry.SubmissionWindowOpened(matchId, expectedWindowClosesAt);
        resultRegistry.signalMatchEnd(matchId);

        (ResultRegistry.MatchStatus status, uint256 windowClosesAt, uint8 attemptNumber,,,) =
            resultRegistry.results(matchId);

        assertEq(uint256(status), uint256(ResultRegistry.MatchStatus.OPEN));
        assertEq(windowClosesAt, expectedWindowClosesAt);
        assertEq(attemptNumber, 1);
    }

    function test_signalMatchEnd_works_for_different_matches() public {
        // Register first match
        _registerMatch();

        // Register second match (different journey)
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](1);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: 2,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });
        matchRegistry.registerBatch(inputs);

        bytes32 matchId2 =
            keccak256(abi.encodePacked(COMPETITION_ID, SEASON_YEAR, uint8(2), HOME_TEAM_ID, AWAY_TEAM_ID));

        // Signal first match
        vm.prank(owner);
        resultRegistry.signalMatchEnd(matchId);

        // Signal second match
        vm.prank(owner);
        resultRegistry.signalMatchEnd(matchId2);

        // Verify both are open
        (ResultRegistry.MatchStatus status1,,,,,) = resultRegistry.results(matchId);
        (ResultRegistry.MatchStatus status2,,,,,) = resultRegistry.results(matchId2);

        assertEq(uint256(status1), uint256(ResultRegistry.MatchStatus.OPEN));
        assertEq(uint256(status2), uint256(ResultRegistry.MatchStatus.OPEN));
    }

    // --- Helper Functions ---

    function _registerMatch() internal {
        MatchRegistry.MatchInput[] memory inputs = new MatchRegistry.MatchInput[](1);
        inputs[0] = MatchRegistry.MatchInput({
            competitionId: COMPETITION_ID,
            seasonStartYear: SEASON_YEAR,
            journeyNumber: JOURNEY,
            homeTeamId: HOME_TEAM_ID,
            awayTeamId: AWAY_TEAM_ID
        });
        matchRegistry.registerBatch(inputs);
    }
}
