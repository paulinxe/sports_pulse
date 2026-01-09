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

    // Public state variables for testing
    CompetitionRegistry public competitionRegistry;
    TeamRegistry public teamRegistry;
    MatchRegistry public matchRegistry;

    function run() external {
        uint256 deployerPrivateKey = vm.envUint("DEPLOYER_PRIVATE_KEY");

        // Get authorized signer address for MatchRegistry
        // This address needs to be the public key derived from the PRIVATE_KEY env var used in the Signer service
        address authorizedSigner = vm.envAddress("AUTHORIZED_SIGNER_ADDRESS");

        // Get contracts owner address
        address contractsOwner = vm.envAddress("CONTRACTS_OWNER_ADDRESS");

        vm.startBroadcast(deployerPrivateKey);

        // Step 1: Deploy CompetitionRegistry
        console.log("Deploying CompetitionRegistry...");
        string[] memory competitionNames = loadCompetitions();
        competitionRegistry = new CompetitionRegistry(competitionNames);
        console.log("CompetitionRegistry deployed at:", address(competitionRegistry));
        console.log("Number of competitions:", competitionNames.length);
        competitionRegistry.transferOwnership(contractsOwner);
        console.log("CompetitionRegistry ownership transferred to:", contractsOwner);

        // Step 2: Deploy TeamRegistry
        console.log("Deploying TeamRegistry...");
        string[] memory teamNames = loadTeams();
        teamRegistry = new TeamRegistry(teamNames);
        console.log("TeamRegistry deployed at:", address(teamRegistry));
        console.log("Number of teams:", teamNames.length);
        teamRegistry.transferOwnership(contractsOwner);
        console.log("TeamRegistry ownership transferred to:", contractsOwner);

        // Step 3: Deploy MatchRegistry (depends on the above two)
        console.log("Deploying MatchRegistry...");
        matchRegistry = new MatchRegistry(
            authorizedSigner,
            competitionRegistry,
            teamRegistry
        );
        console.log("MatchRegistry deployed at:", address(matchRegistry));
        console.log("Authorized signer:", authorizedSigner);
        matchRegistry.transferOwnership(contractsOwner);
        console.log("MatchRegistry ownership transferred to:", contractsOwner);

        vm.stopBroadcast();

        // Log deployment summary
        console.log("\n=== Deployment Summary ===");
        console.log("CompetitionRegistry:", address(competitionRegistry));
        console.log("TeamRegistry:", address(teamRegistry));
        console.log("MatchRegistry:", address(matchRegistry));
        console.log("Authorized Signer:", authorizedSigner);
        console.log("Contracts Owner:", contractsOwner);
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

