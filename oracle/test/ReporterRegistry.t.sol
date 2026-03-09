// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";
import {ReporterRegistry} from "../src/ReporterRegistry.sol";
import {ConsensusEngine} from "../src/ConsensusEngine.sol";

contract ReporterRegistryTest is Test {
    ReporterRegistry public registry;
    ConsensusEngine public consensusEngine;

    address public reporter1;
    address public reporter2;
    address public reporter3;

    function setUp() public {
        consensusEngine = new ConsensusEngine();
        registry = new ReporterRegistry(address(consensusEngine));
        reporter1 = makeAddr("reporter1");
        reporter2 = makeAddr("reporter2");
        reporter3 = makeAddr("reporter3");
    }

    // --- stake() ---

    function test_stake_reverts_when_zero_value() public {
        vm.prank(reporter1);
        vm.expectRevert(ReporterRegistry.ZeroStake.selector);
        registry.stake{value: 0}();
    }

    function test_stake_updates_balance_and_emits() public {
        uint256 amount = 1 ether;
        vm.deal(reporter1, amount);

        vm.prank(reporter1);
        vm.expectEmit(true, true, true, true);
        emit ReporterRegistry.Staked(reporter1, amount);
        registry.stake{value: amount}();

        (uint256 staked, uint256 rewards, uint256 requestedAt, uint256 correct, uint256 incorrect) =
            registry.reporters(reporter1);
        assertEq(staked, amount);
        assertEq(requestedAt, 0);
        assertEq(rewards, 0);
        assertEq(correct, 0);
        assertEq(incorrect, 0);
    }

    function test_stake_top_up_adds_to_balance() public {
        vm.deal(reporter1, 2 ether);

        vm.startPrank(reporter1);
        registry.stake{value: 1 ether}();
        (uint256 stakedAfterFirst,,,,) = registry.reporters(reporter1);
        assertEq(stakedAfterFirst, 1 ether);
        registry.stake{value: 0.5 ether}();
        (uint256 staked,,,,) = registry.reporters(reporter1);
        assertEq(staked, 1.5 ether);
        vm.stopPrank();
    }

    function test_stake_multiple_reporters_independent() public {
        vm.deal(reporter1, 1 ether);
        vm.deal(reporter2, 2 ether);

        vm.prank(reporter1);
        registry.stake{value: 1 ether}();
        vm.prank(reporter2);
        registry.stake{value: 2 ether}();

        (uint256 s1,,,,) = registry.reporters(reporter1);
        (uint256 s2,,,,) = registry.reporters(reporter2);
        assertEq(s1, 1 ether);
        assertEq(s2, 2 ether);
    }

    // --- requestWithdrawal() ---

    function test_requestWithdrawal_reverts_when_no_staked_balance() public {
        vm.prank(reporter1);
        vm.expectRevert(ReporterRegistry.NoStakedBalance.selector);
        registry.requestWithdrawal();
    }

    function test_requestWithdrawal_reverts_when_already_requested() public {
        vm.deal(reporter1, 1 ether);
        vm.startPrank(reporter1);
        registry.stake{value: 1 ether}();
        registry.requestWithdrawal();
        vm.expectRevert(ReporterRegistry.WithdrawalAlreadyRequested.selector);
        registry.requestWithdrawal();
        vm.stopPrank();
    }

    function test_requestWithdrawal_sets_claimableAt_and_emits() public {
        vm.deal(reporter1, 1 ether);
        uint256 requestTime = block.timestamp;
        uint256 claimableAt = requestTime + registry.WITHDRAWAL_COOLDOWN();

        vm.startPrank(reporter1);
        registry.stake{value: 1 ether}();
        vm.expectEmit(true, true, true, true);
        emit ReporterRegistry.WithdrawalRequested(reporter1, claimableAt);
        registry.requestWithdrawal();
        vm.stopPrank();

        (uint256 staked,, uint256 requestedAt,,) = registry.reporters(reporter1);
        assertEq(staked, 1 ether);
        assertEq(requestedAt, requestTime);
    }

    // --- withdraw() ---

    function test_withdraw_reverts_when_not_requested() public {
        vm.deal(reporter1, 1 ether);
        vm.startPrank(reporter1);
        registry.stake{value: 1 ether}();
        vm.expectRevert(ReporterRegistry.WithdrawalNotRequested.selector);
        registry.withdraw();
        vm.stopPrank();
    }

    function test_withdraw_reverts_when_cooldown_not_elapsed() public {
        vm.deal(reporter1, 1 ether);
        vm.startPrank(reporter1);
        registry.stake{value: 1 ether}();
        registry.requestWithdrawal();
        vm.stopPrank();

        (,, uint256 requestedAt,,) = registry.reporters(reporter1);
        uint256 claimableAt = requestedAt + registry.WITHDRAWAL_COOLDOWN();
        vm.warp(claimableAt - 1);
        vm.prank(reporter1);
        vm.expectRevert(abi.encodeWithSelector(ReporterRegistry.CooldownNotElapsed.selector, claimableAt));
        registry.withdraw();
    }

    function test_withdraw_after_cooldown_transfers_eth_and_clears_balance() public {
        vm.deal(reporter1, 1 ether);
        vm.startPrank(reporter1);
        registry.stake{value: 1 ether}();
        registry.requestWithdrawal();
        vm.stopPrank();

        (,, uint256 requestedAt,,) = registry.reporters(reporter1);
        uint256 claimableAt = requestedAt + registry.WITHDRAWAL_COOLDOWN();
        vm.warp(claimableAt);

        uint256 balanceBefore = reporter1.balance;
        vm.startPrank(reporter1);
        vm.expectEmit(true, true, true, true);
        emit ReporterRegistry.Withdrawn(reporter1, 1 ether);
        registry.withdraw();
        vm.stopPrank();

        assertEq(reporter1.balance, balanceBefore + 1 ether);
        (uint256 staked, uint256 rewards, uint256 reqAt, uint256 correct, uint256 incorrect) =
            registry.reporters(reporter1);
        assertEq(staked, 0);
        assertEq(reqAt, 0);
        assertEq(rewards, 0);
        assertEq(correct, 0);
        assertEq(incorrect, 0);
    }

    function test_withdraw_exactly_at_cooldown_succeeds() public {
        vm.deal(reporter1, 1 ether);
        vm.startPrank(reporter1);
        registry.stake{value: 1 ether}();
        registry.requestWithdrawal();
        vm.stopPrank();

        vm.warp(block.timestamp + registry.WITHDRAWAL_COOLDOWN());
        vm.prank(reporter1);
        registry.withdraw();
        assertEq(reporter1.balance, 1 ether);
        (uint256 staked,,,,) = registry.reporters(reporter1);
        assertEq(staked, 0);
    }

    // --- isEligible() ---

    function test_isEligible_returns_false_for_unknown_address() public {
        assertFalse(registry.isEligible(reporter1));
        assertFalse(registry.isEligible(address(0)));
        assertFalse(registry.isEligible(makeAddr("random")));
    }

    function test_isEligible_returns_false_when_staked_below_minimum() public {
        uint256 belowMin = registry.MIN_STAKE() - 1;
        vm.deal(reporter1, belowMin);
        vm.prank(reporter1);
        registry.stake{value: belowMin}();
        assertFalse(registry.isEligible(reporter1));
    }

    function test_isEligible_returns_true_when_staked_exactly_minimum() public {
        uint256 minStake = registry.MIN_STAKE();
        vm.deal(reporter1, minStake);
        vm.prank(reporter1);
        registry.stake{value: minStake}();
        assertTrue(registry.isEligible(reporter1));
    }

    function test_isEligible_returns_true_when_staked_above_minimum() public {
        uint256 aboveMin = registry.MIN_STAKE() + 1;
        vm.deal(reporter1, aboveMin);
        vm.prank(reporter1);
        registry.stake{value: aboveMin}();
        assertTrue(registry.isEligible(reporter1));
    }

    function test_isEligible_returns_true_when_pending_withdrawal_still_eligible() public {
        uint256 minStake = registry.MIN_STAKE();
        vm.deal(reporter1, minStake);
        vm.startPrank(reporter1);
        registry.stake{value: minStake}();
        registry.requestWithdrawal();
        vm.stopPrank();
        // During cooldown, ETH is still locked and balance >= MIN_STAKE
        assertTrue(registry.isEligible(reporter1));
    }

    function test_isEligible_returns_false_after_withdraw() public {
        uint256 minStake = registry.MIN_STAKE();
        vm.deal(reporter1, minStake);
        vm.startPrank(reporter1);
        registry.stake{value: minStake}();
        registry.requestWithdrawal();
        vm.stopPrank();
        vm.warp(block.timestamp + registry.WITHDRAWAL_COOLDOWN());
        vm.prank(reporter1);
        registry.withdraw();
        assertFalse(registry.isEligible(reporter1));
    }

    function test_isEligible_works_with_multiple_reporters_independent() public {
        uint256 minStake = registry.MIN_STAKE();
        vm.deal(reporter1, minStake);
        vm.deal(reporter2, minStake - 1);
        vm.prank(reporter1);
        registry.stake{value: minStake}();
        vm.prank(reporter2);
        registry.stake{value: minStake - 1}();
        assertTrue(registry.isEligible(reporter1));
        assertFalse(registry.isEligible(reporter2));
    }

    // --- claimSlashedRewards() ---

    function test_claimSlashedRewards_reverts_when_nothing_to_claim() public {
        vm.prank(reporter1);
        vm.expectRevert(ReporterRegistry.NothingToClaim.selector);
        registry.claimSlashedRewards();
    }

    function test_claimSlashedRewards_reverts_when_reporter_has_stake_but_no_rewards() public {
        vm.deal(reporter1, 1 ether);
        vm.startPrank(reporter1);
        registry.stake{value: 1 ether}();
        vm.expectRevert(ReporterRegistry.NothingToClaim.selector);
        registry.claimSlashedRewards();
        vm.stopPrank();
    }

    // --- slash() ---

    function test_constructor_reverts_when_consensus_engine_zero() public {
        vm.expectRevert(ReporterRegistry.ConsensusEngineZero.selector);
        new ReporterRegistry(address(0));
    }

    function test_slash_reverts_when_caller_is_not_consensus_engine() public {
        vm.deal(reporter1, 1 ether);
        vm.prank(reporter1);
        registry.stake{value: 1 ether}();

        address[] memory wrong = new address[](1);
        wrong[0] = reporter1;
        address[] memory correct;
        vm.prank(reporter1);
        vm.expectRevert(ReporterRegistry.OnlyConsensusEngine.selector);
        registry.slash(wrong, correct);
    }

    function test_slash_reverts_when_exceeds_max_reporters() public {
        address[] memory wrong = new address[](3);
        wrong[0] = reporter1;
        wrong[1] = reporter2;
        wrong[2] = reporter3;
        address[] memory correct = new address[](3);
        correct[0] = makeAddr("c1");
        correct[1] = makeAddr("c2");
        correct[2] = makeAddr("c3");
        vm.prank(address(consensusEngine));
        vm.expectRevert(ReporterRegistry.ExceedsMaxReporters.selector);
        registry.slash(wrong, correct);
    }

    function test_slash_one_wrong_reduces_balance() public {
        vm.deal(reporter1, 1 ether);
        vm.prank(reporter1);
        registry.stake{value: 1 ether}();

        address[] memory wrong = new address[](1);
        wrong[0] = reporter1;
        address[] memory correct = new address[](1);
        correct[0] = reporter2;
        vm.prank(address(consensusEngine));
        vm.expectEmit(true, true, true, true);
        emit ReporterRegistry.Slashed(reporter1, 0.25 ether);
        registry.slash(wrong, correct);

        (uint256 staked,,,,) = registry.reporters(reporter1);
        assertEq(staked, 0.75 ether);
    }

    function test_slash_one_wrong_one_correct_distributes_rewards() public {
        vm.deal(reporter1, 1 ether);
        vm.deal(reporter2, 1 ether);
        vm.prank(reporter1);
        registry.stake{value: 1 ether}();
        vm.prank(reporter2);
        registry.stake{value: 1 ether}();

        address[] memory wrong = new address[](1);
        wrong[0] = reporter1;
        address[] memory correct = new address[](1);
        correct[0] = reporter2;
        vm.prank(address(consensusEngine));
        registry.slash(wrong, correct);

        (uint256 staked1,,,,) = registry.reporters(reporter1);
        (, uint256 rewards2,,,) = registry.reporters(reporter2);
        assertEq(staked1, 0.75 ether);
        assertEq(rewards2, 0.25 ether);

        vm.prank(reporter2);
        registry.claimSlashedRewards();
        assertEq(reporter2.balance, 0.25 ether, "reporter2 received slashed rewards");
        (, uint256 rewardsAfter,,,) = registry.reporters(reporter2);
        assertEq(rewardsAfter, 0);
    }

    function test_slash_two_wrong_two_correct_divides_equally_remainder_stays() public {
        vm.deal(reporter1, 1 ether);
        vm.deal(reporter2, 1 ether);
        vm.deal(reporter3, 1 ether);
        address reporter4 = makeAddr("reporter4");
        vm.deal(reporter4, 1 ether);
        vm.prank(reporter1);
        registry.stake{value: 1 ether}();
        vm.prank(reporter2);
        registry.stake{value: 1 ether}();
        vm.prank(reporter3);
        registry.stake{value: 1 ether}();
        vm.prank(reporter4);
        registry.stake{value: 1 ether}();

        address[] memory wrong = new address[](2);
        wrong[0] = reporter1;
        wrong[1] = reporter2;
        address[] memory correct = new address[](2);
        correct[0] = reporter3;
        correct[1] = reporter4;
        vm.prank(address(consensusEngine));
        registry.slash(wrong, correct);

        // 0.25 + 0.25 = 0.5 ether slashed. rewardShare = 0.5/2 = 0.25 each.
        (uint256 staked1,,,,) = registry.reporters(reporter1);
        (uint256 staked2,,,,) = registry.reporters(reporter2);
        (, uint256 rewards3,,,) = registry.reporters(reporter3);
        (, uint256 rewards4,,,) = registry.reporters(reporter4);
        assertEq(staked1, 0.75 ether);
        assertEq(staked2, 0.75 ether);
        assertEq(rewards3, 0.25 ether);
        assertEq(rewards4, 0.25 ether);
    }

    function test_slash_reverts_when_zero_correct_reporters() public {
        vm.deal(reporter1, 1 ether);
        vm.prank(reporter1);
        registry.stake{value: 1 ether}();

        address[] memory wrong = new address[](1);
        wrong[0] = reporter1;
        address[] memory correct;
        vm.prank(address(consensusEngine));
        vm.expectRevert(ReporterRegistry.ZeroCorrectReporters.selector);
        registry.slash(wrong, correct);
    }

    function test_slash_skips_wrong_reporter_with_zero_stake() public {
        // reporter1 was completely slashed; slash() continues, we simply don't slash (0 amount)
        address[] memory wrong = new address[](1);
        wrong[0] = reporter1;
        (uint256 staked,,,,) = registry.reporters(reporter1);
        assertEq(staked, 0);
        address[] memory correct = new address[](1);
        correct[0] = reporter2;
        vm.prank(address(consensusEngine));
        registry.slash(wrong, correct);
        (staked,,,,) = registry.reporters(reporter1);
        assertEq(staked, 0);
    }

    function test_slash_skips_zero_address_in_wrong_reporters() public {
        address[] memory wrong = new address[](1);
        wrong[0] = address(0);
        address[] memory correct = new address[](1);
        correct[0] = reporter1;
        vm.prank(address(consensusEngine));
        registry.slash(wrong, correct);
        // No revert; totalSlashed is 0 so no distribution
    }

    function test_slash_integer_division_remainder_stays_in_contract() public {
        // 4 wei * 25% = 1 wei slashed. 2 correct reporters -> rewardShare = 1/2 = 0 each. Remainder 1 wei stays in contract.
        vm.deal(reporter1, 4);
        vm.prank(reporter1);
        registry.stake{value: 4}();
        address[] memory wrong = new address[](1);
        wrong[0] = reporter1;
        address[] memory correct = new address[](2);
        correct[0] = reporter2;
        correct[1] = reporter3;
        vm.prank(address(consensusEngine));
        registry.slash(wrong, correct);
        (uint256 stakedWrong,,,,) = registry.reporters(reporter1);
        (, uint256 reporter2Rewards,,,) = registry.reporters(reporter2);
        (, uint256 reporter3Rewards,,,) = registry.reporters(reporter3);
        assertEq(stakedWrong, 3, "wrong reporter lost 1 wei");
        assertEq(reporter2Rewards, 0, "rewardShare 1 wei / 2 = 0");
        assertEq(reporter3Rewards, 0, "rewardShare 1 wei / 2 = 0");
        assertEq(reporter2Rewards + reporter3Rewards, 0, "remainder stays in contract (dust)");
    }

    function test_slash_below_min_loses_eligibility() public {
        uint256 minStake = registry.MIN_STAKE();
        vm.deal(reporter1, minStake);
        vm.prank(reporter1);
        registry.stake{value: minStake}();
        assertTrue(registry.isEligible(reporter1));

        address[] memory wrong = new address[](1);
        wrong[0] = reporter1;
        address[] memory correct = new address[](1);
        correct[0] = reporter2;
        vm.prank(address(consensusEngine));
        registry.slash(wrong, correct);
        // MIN_STAKE (0.1 ether) * 25% = 0.025, new balance 0.075 ether < 0.1
        (uint256 staked,,,,) = registry.reporters(reporter1);
        assertEq(staked, 0.075 ether);
        assertFalse(registry.isEligible(reporter1));
    }
}
