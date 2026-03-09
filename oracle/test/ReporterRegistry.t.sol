// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Test} from "forge-std/Test.sol";
import {ReporterRegistry} from "../src/ReporterRegistry.sol";

contract ReporterRegistryTest is Test {
    ReporterRegistry public registry;

    address public reporter1;
    address public reporter2;

    function setUp() public {
        registry = new ReporterRegistry();
        reporter1 = makeAddr("reporter1");
        reporter2 = makeAddr("reporter2");
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

        (uint256 staked, , uint256 requestedAt, , ) = registry.reporters(reporter1);
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

        (, , uint256 requestedAt, , ) = registry.reporters(reporter1);
        uint256 claimableAt = requestedAt + registry.WITHDRAWAL_COOLDOWN();
        vm.warp(claimableAt - 1);
        vm.prank(reporter1);
        vm.expectRevert(
            abi.encodeWithSelector(ReporterRegistry.CooldownNotElapsed.selector, claimableAt)
        );
        registry.withdraw();
    }

    function test_withdraw_after_cooldown_transfers_eth_and_clears_balance() public {
        vm.deal(reporter1, 1 ether);
        vm.startPrank(reporter1);
        registry.stake{value: 1 ether}();
        registry.requestWithdrawal();
        vm.stopPrank();

        (, , uint256 requestedAt, , ) = registry.reporters(reporter1);
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

    // TODO: we miss tests for the case when the reporter has rewards and tries to claim them
    // This will come when tackling the Slashing part
}
