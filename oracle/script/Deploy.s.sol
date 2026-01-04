// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Script} from "forge-std/Script.sol";
import {stdJson} from "forge-std/StdJson.sol";
import {console} from "forge-std/console.sol";
import {CompetitionRegistry} from "../src/CompetitionRegistry.sol";
import {TeamRegistry} from "../src/TeamRegistry.sol";
import {MatchRegistry} from "../src/MatchRegistry.sol";

contract Deploy is Script {
    using stdJson for string;

    // Paths to JSON files
    string constant COMPETITIONS_JSON = "./script/data/competitions.json";
    string constant TEAMS_JSON = "./script/data/teams.json";

    function run() external {
        // At the moment, using one from Anvil
        uint256 deployerPrivateKey = 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80;
        
        // Get authorized signer address for MatchRegistry
        // This address needs to be the public key derived from the PRIVATE_KEY env var used in the Signer service
        address authorizedSigner = 0xCC0724CDc18DaE6B469b8e8B533fCd4dBE32FB46;

        vm.startBroadcast(deployerPrivateKey);

        // Step 1: Deploy CompetitionRegistry
        console.log("Deploying CompetitionRegistry...");
        string[] memory competitionNames = loadCompetitions();
        CompetitionRegistry competitionRegistry = new CompetitionRegistry(competitionNames);
        console.log("CompetitionRegistry deployed at:", address(competitionRegistry));
        console.log("Number of competitions:", competitionNames.length);

        // Step 2: Deploy TeamRegistry
        console.log("Deploying TeamRegistry...");
        string[] memory teamNames = loadTeams();
        TeamRegistry teamRegistry = new TeamRegistry(teamNames);
        console.log("TeamRegistry deployed at:", address(teamRegistry));
        console.log("Number of teams:", teamNames.length);

        // Step 3: Deploy MatchRegistry (depends on the above two)
        console.log("Deploying MatchRegistry...");
        MatchRegistry matchRegistry = new MatchRegistry(
            authorizedSigner,
            competitionRegistry,
            teamRegistry
        );
        console.log("MatchRegistry deployed at:", address(matchRegistry));
        console.log("Authorized signer:", authorizedSigner);

        vm.stopBroadcast();

        // Log deployment summary
        console.log("\n=== Deployment Summary ===");
        console.log("CompetitionRegistry:", address(competitionRegistry));
        console.log("TeamRegistry:", address(teamRegistry));
        console.log("MatchRegistry:", address(matchRegistry));
        console.log("Authorized Signer:", authorizedSigner);
    }

    function loadCompetitions() internal view returns (string[] memory) {
        string memory json = vm.readFile(COMPETITIONS_JSON);
        return json.readStringArray(".competitions");
    }

    function loadTeams() internal view returns (string[] memory) {
        string memory json = vm.readFile(TEAMS_JSON);
        return json.readStringArray(".teams");
    }
}

