// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Script} from "forge-std/Script.sol";
import {stdJson} from "forge-std/StdJson.sol";
import {console} from "forge-std/console.sol";
import {CompetitionRegistry} from "../src/CompetitionRegistry.sol";
import {TeamRegistry} from "../src/TeamRegistry.sol";
import {MatchRegistry} from "../src/MatchRegistry.sol";

// This script deploys contracts separately. This means this is NOT atomic.
// Is not an issue in our case as we do need to transfer ownership after deployment
// but this can wait for some blocks to be processed.
contract Deploy is Script {
    using stdJson for string;

    // Paths to JSON files
    string constant COMPETITIONS_JSON = "./script/data/competitions.json";
    string constant TEAMS_JSON = "./script/data/teams.json";

    // Public state variables for testing
    CompetitionRegistry public competitionRegistry;
    TeamRegistry public teamRegistry;
    MatchRegistry public matchRegistry;

    /// @notice Split in this way so we can test this script without environment issues
    function deploy(
        uint256 deployerPrivateKey,
        address authorizedSigner,
        address contractsOwner
    ) public virtual {
        // Security validations
        require(authorizedSigner != address(0), "AUTHORIZED_SIGNER_ADDRESS cannot be zero");
        require(contractsOwner != address(0), "CONTRACTS_OWNER_ADDRESS cannot be zero");
        require(authorizedSigner != contractsOwner, "AUTHORIZED_SIGNER_ADDRESS and CONTRACTS_OWNER_ADDRESS must be distinct");

        address deployer = vm.addr(deployerPrivateKey);
        require(deployer != authorizedSigner, "Deployer and AUTHORIZED_SIGNER_ADDRESS must be distinct");
        require(deployer != contractsOwner, "Deployer and CONTRACTS_OWNER_ADDRESS must be distinct");

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

    /// @notice Parameterless run function that reads from environment variables
    /// @dev This is the entry point for Makefile usage
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("DEPLOYER_PRIVATE_KEY");
        address authorizedSigner = vm.envAddress("AUTHORIZED_SIGNER_ADDRESS");
        address contractsOwner = vm.envAddress("CONTRACTS_OWNER_ADDRESS");

        deploy(deployerPrivateKey, authorizedSigner, contractsOwner);
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

