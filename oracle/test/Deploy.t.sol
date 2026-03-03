// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

/*
import {Test} from "forge-std/Test.sol";
import {Deploy} from "../script/Deploy.s.sol";
import {CompetitionRegistry} from "../src/CompetitionRegistry.sol";
import {TeamRegistry} from "../src/TeamRegistry.sol";
import {MatchRegistry} from "../src/MatchRegistry.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

contract DeployTest is Test {
    Deploy public deployScript;
    address public deployer;
    address public authorizedSigner;
    address public contractsOwner;
    uint256 public deployerPrivateKey;

    CompetitionRegistry public competitionRegistry;
    TeamRegistry public teamRegistry;
    MatchRegistry public matchRegistry;

    function setUp() public {
        deployerPrivateKey = uint256(0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80); // Anvil default
        deployer = vm.addr(deployerPrivateKey);
        authorizedSigner = makeAddr("authorized_signer");
        contractsOwner = makeAddr("contracts_owner");
    }

    function test_deployment_reverts_if_authorized_signer_is_zero() public {
        Deploy _deployScript = new Deploy();

        vm.expectRevert("AUTHORIZED_SIGNER_ADDRESS cannot be zero");

        _deployScript.deploy(deployerPrivateKey, address(0), contractsOwner);
    }

    function test_deployment_reverts_if_contracts_owner_is_zero() public {
        Deploy _deployScript = new Deploy();

        vm.expectRevert("CONTRACTS_OWNER_ADDRESS cannot be zero");

        _deployScript.deploy(deployerPrivateKey, authorizedSigner, address(0));
    }

    function test_deployment_reverts_if_authorized_signer_equals_contracts_owner() public {
        address sameAddress = makeAddr("same_address");
        Deploy _deployScript = new Deploy();

        vm.expectRevert("AUTHORIZED_SIGNER_ADDRESS and CONTRACTS_OWNER_ADDRESS must be distinct");

        _deployScript.deploy(deployerPrivateKey, sameAddress, sameAddress);
    }

    function test_deployment_reverts_if_deployer_equals_authorized_signer() public {
        Deploy _deployScript = new Deploy();

        vm.expectRevert("Deployer and AUTHORIZED_SIGNER_ADDRESS must be distinct");

        _deployScript.deploy(deployerPrivateKey, deployer, contractsOwner);
    }

    function test_deployment_reverts_if_deployer_equals_contracts_owner() public {
        Deploy _deployScript = new Deploy();

        vm.expectRevert("Deployer and CONTRACTS_OWNER_ADDRESS must be distinct");

        _deployScript.deploy(deployerPrivateKey, authorizedSigner, deployer);
    }

    function test_deployment_script_deploys_all_contracts() public {
        _deployAndVerify();

        assertGt(address(competitionRegistry).code.length, 0, "CompetitionRegistry should have code");
        assertGt(address(teamRegistry).code.length, 0, "TeamRegistry should have code");
        assertGt(address(matchRegistry).code.length, 0, "MatchRegistry should have code");
    }

    function test_contracts_ownership_is_transferred() public {
        _deployAndVerify();

        assertEq(competitionRegistry.owner(), contractsOwner, "CompetitionRegistry owner should be contractsOwner");
        assertEq(teamRegistry.owner(), contractsOwner, "TeamRegistry owner should be contractsOwner");
        assertEq(matchRegistry.owner(), contractsOwner, "MatchRegistry owner should be contractsOwner");
    }

    function test_competition_registry_is_initialized_with_data() public {
        _deployAndVerify();

        assertEq(competitionRegistry.competitions(1), "LaLiga", "First competition should be LaLiga");
        assertEq(competitionRegistry.competitionIdCounter(), 1, "Should have 1 competition");
    }

    function test_team_registry_is_initialized_with_data() public {
        _deployAndVerify();

        assertEq(teamRegistry.teams(1), "Alaves", "First team should be Alaves");
        assertEq(teamRegistry.teams(2), "Athletic Club", "Second team should be Athletic Club");
        assertEq(teamRegistry.teamIdCounter(), 20, "Should have 20 teams");
    }

    function test_match_registry_is_initialized_correctly() public {
        _deployAndVerify();

        assertEq(matchRegistry.authorizedSigner(), authorizedSigner, "Authorized signer should match");
        assertEq(address(matchRegistry.COMPETITION_REGISTRY()), address(competitionRegistry), "CompetitionRegistry reference should match");
        assertEq(address(matchRegistry.TEAM_REGISTRY()), address(teamRegistry), "TeamRegistry reference should match");
    }

    function test_only_owner_can_add_competitions_after_deployment() public {
        _deployAndVerify();

        address nonOwner = makeAddr("non_owner");
        vm.prank(nonOwner);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, nonOwner));
        competitionRegistry.addCompetition("PremierLeague");

        vm.prank(contractsOwner);
        competitionRegistry.addCompetition("PremierLeague");
        assertEq(competitionRegistry.competitions(2), "PremierLeague", "Owner should be able to add competition");
    }

    function test_only_owner_can_add_teams_after_deployment() public {
        _deployAndVerify();

        address nonOwner = makeAddr("non_owner");
        string[] memory newTeams = new string[](1);
        newTeams[0] = "New Team";

        vm.prank(nonOwner);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, nonOwner));
        teamRegistry.addTeams(newTeams);

        vm.prank(contractsOwner);
        teamRegistry.addTeams(newTeams);
        assertEq(teamRegistry.teams(21), "New Team", "Owner should be able to add teams");
    }

    function test_only_owner_can_rotate_signer_after_deployment() public {
        _deployAndVerify();

        address nonOwner = makeAddr("non_owner");
        address newSigner = makeAddr("new_signer");

        vm.prank(nonOwner);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, nonOwner));
        matchRegistry.rotateSigner(newSigner);

        vm.prank(contractsOwner);
        matchRegistry.rotateSigner(newSigner);
        assertEq(matchRegistry.authorizedSigner(), newSigner, "Owner should be able to rotate signer");
    }

    function _deployAndVerify() private {
        deployScript = new Deploy();
        deployScript.deploy(deployerPrivateKey, authorizedSigner, contractsOwner);

        competitionRegistry = deployScript.competitionRegistry();
        teamRegistry = deployScript.teamRegistry();
        matchRegistry = deployScript.matchRegistry();

        assertNotEq(address(competitionRegistry), address(0), "CompetitionRegistry should be deployed");
        assertNotEq(address(teamRegistry), address(0), "TeamRegistry should be deployed");
        assertNotEq(address(matchRegistry), address(0), "MatchRegistry should be deployed");
    }
}
*/
