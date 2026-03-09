// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";
import {TeamRegistry} from "../src/TeamRegistry.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {Strings} from "@openzeppelin/contracts/utils/Strings.sol";

contract TeamRegistryTest is Test {
    TeamRegistry public teamRegistry;

    function setUp() public {
        string[] memory teamNames = new string[](1);
        teamNames[0] = "Nacional";

        teamRegistry = new TeamRegistry(teamNames);
    }

    function test_registry_is_initialized_with_the_correct_teams() public view {
        assertEq(teamRegistry.teams(1), "Nacional");
    }

    function test_tx_reverts_if_not_owner() public {
        address notOwner = makeAddr("not_owner");
        vm.prank(notOwner);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, notOwner));
        string[] memory teamNames = new string[](1);
        teamNames[0] = "Albion";
        teamRegistry.addTeams(teamNames);
    }

    function test_tx_reverts_if_too_many_teams_on_creation() public {
        string[] memory teamNames = new string[](201);
        for (uint8 i = 0; i < 201; i++) {
            teamNames[i] = string.concat("Team ", Strings.toString(i));
        }

        vm.expectRevert(abi.encodeWithSelector(TeamRegistry.TooManyTeams.selector, 201));
        new TeamRegistry(teamNames);
    }

    function test_tx_reverts_if_team_name_is_empty() public {
        vm.expectRevert(abi.encodeWithSelector(TeamRegistry.InvalidTeamName.selector));
        string[] memory teamNames = new string[](1);
        teamNames[0] = "";
        teamRegistry.addTeams(teamNames);
    }

    function test_we_can_add_a_team() public {
        string[] memory teamNames = new string[](2);
        teamNames[0] = "Boca Juniors";
        teamNames[1] = "River Plate";

        vm.expectEmit(true, true, true, true);
        emit TeamRegistry.TeamAdded(2, "Boca Juniors");
        vm.expectEmit(true, true, true, true);
        emit TeamRegistry.TeamAdded(3, "River Plate");

        teamRegistry.addTeams(teamNames);

        assertEq(teamRegistry.teams(2), "Boca Juniors");
        assertEq(teamRegistry.teams(3), "River Plate");
    }
}
