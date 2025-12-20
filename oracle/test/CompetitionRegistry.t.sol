// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import { Test } from "forge-std/Test.sol";
import { CompetitionRegistry } from "../src/CompetitionRegistry.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";

contract CompetitionRegistryTest is Test {
    CompetitionRegistry public competitionRegistry;

    function setUp() public {
        string[] memory competitionNames = new string[](1);
        competitionNames[0] = "LaLiga";

        competitionRegistry = new CompetitionRegistry(competitionNames);
    }

    function test_registry_is_initialized_with_the_correct_competitions() public view {
        assertEq(competitionRegistry.competitions(1), "LaLiga");
    }

    function test_tx_reverts_if_not_owner() public {
        address notOwner = makeAddr("not_owner");
        vm.prank(notOwner);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, notOwner));
        competitionRegistry.addCompetition("PremierLeague");
    }

    function test_tx_reverts_if_too_many_competitions_on_creation() public {
        string[] memory competitionNames = new string[](201);
        for (uint8 i = 0; i < 201; i++) {
            competitionNames[i] = string(abi.encodePacked("Competition ", Strings.toString(i)));
        }

        vm.expectRevert(abi.encodeWithSelector(CompetitionRegistry.TooManyCompetitions.selector, 201));
        new CompetitionRegistry(competitionNames);
    }

    function test_we_can_add_a_league() public {
        vm.expectEmit(true, true, true, true);
        emit CompetitionRegistry.CompetitionAdded(2, "PremierLeague");
        competitionRegistry.addCompetition("PremierLeague");
        assertEq(competitionRegistry.competitions(2), "PremierLeague");
    }
}