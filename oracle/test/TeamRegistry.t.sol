// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import { Test } from "forge-std/Test.sol";
import { TeamRegistry } from "../src/TeamRegistry.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";

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
        teamRegistry.addTeam("Albion");
    }

    function test_tx_reverts_if_too_many_teams_on_creation() public {
        string[] memory teamNames = new string[](201);
        for (uint8 i = 0; i < 201; i++) {
            teamNames[i] = string(abi.encodePacked("Team ", Strings.toString(i)));
        }

        vm.expectRevert(abi.encodeWithSelector(TeamRegistry.TooManyTeams.selector, 201));
        new TeamRegistry(teamNames);
    }
    
    function test_we_can_add_a_team() public {
        vm.expectEmit(true, true, true, true);
        emit TeamRegistry.TeamAdded(2, "Boca Juniors");
        teamRegistry.addTeam("Boca Juniors");
        assertEq(teamRegistry.teams(2), "Boca Juniors");
    }
}