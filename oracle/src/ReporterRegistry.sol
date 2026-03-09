// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

/**
 * @title ReporterRegistry
 * @notice Staking and Reporter eligibility. Reporters stake ETH to participate; slashing and rewards are applied here.
 */
contract ReporterRegistry {
    struct Reporter {
        uint256 stakedBalance;          // current staked ETH
        uint256 claimableRewards;       // accumulated slashed ETH from honest submissions
        uint256 withdrawalRequestedAt;  // block.timestamp of withdrawal request, 0 if none
        uint256 correctSubmissions;     // reputation tracking
        uint256 incorrectSubmissions;   // reputation tracking
    }

    mapping(address => Reporter) public reporters;

    uint256 public constant MIN_STAKE = 0.1 ether;
    uint256 public constant WITHDRAWAL_COOLDOWN = 7 days;

    event Staked(address indexed reporter, uint256 amount);
    event WithdrawalRequested(address indexed reporter, uint256 claimableAt);
    event Withdrawn(address indexed reporter, uint256 amount);
    event Slashed(address indexed reporter, uint256 amount);
    event SlashedRewardsClaimed(address indexed reporter, uint256 amount);

    error ZeroStake();
    error NoStakedBalance();
    error WithdrawalAlreadyRequested();
    error WithdrawalNotRequested();
    error CooldownNotElapsed(uint256 claimableAt);
    error NothingToClaim();

    /**
     * @notice Stake ETH. Adds to stakedBalance (first-time or top-up after slash).
     */
    function stake() external payable {
        if (msg.value == 0) {
            revert ZeroStake();
        }

        reporters[msg.sender].stakedBalance += msg.value;
        emit Staked(msg.sender, msg.value);
    }

    /**
     * @notice Start the 7-day withdrawal cooldown. No pending withdrawal must exist.
     */
    function requestWithdrawal() external {
        Reporter storage reporter = reporters[msg.sender];
        if (reporter.stakedBalance == 0) {
            revert NoStakedBalance();
        }

        if (reporter.withdrawalRequestedAt != 0) {
            revert WithdrawalAlreadyRequested();
        }

        reporter.withdrawalRequestedAt = block.timestamp;
        uint256 claimableAt = block.timestamp + WITHDRAWAL_COOLDOWN;
        emit WithdrawalRequested(msg.sender, claimableAt);
    }

    /**
     * @notice Withdraw full staked balance after cooldown.
     */
    function withdraw() external {
        Reporter storage reporter = reporters[msg.sender];
        uint256 withdrawalRequestedAt = reporter.withdrawalRequestedAt;
        if (withdrawalRequestedAt == 0) {
            revert WithdrawalNotRequested();
        }

        uint256 claimableAt = withdrawalRequestedAt + WITHDRAWAL_COOLDOWN;
        if (block.timestamp < claimableAt) {
            revert CooldownNotElapsed(claimableAt);
        }

        uint256 amount = reporter.stakedBalance;
        reporter.stakedBalance = 0;
        reporter.withdrawalRequestedAt = 0;
        (bool ok,) = msg.sender.call{value: amount}("");
        require(ok, "Transfer failed"); // TODO: check here if we need a custom error
        emit Withdrawn(msg.sender, amount);
    }

    /**
     * @notice Check if a reporter is eligible to submit (staked balance >= minimum).
     * @dev Used by Report Aggregator before accepting a submission. Pending withdrawal does not affect eligibility.
     */
    function isEligible(address reporter) external view returns (bool) {
        return reporters[reporter].stakedBalance >= MIN_STAKE;
    }

    /**
     * @notice Claim accumulated slashed rewards. Pull model; resets claimableRewards to 0.
     */
    function claimSlashedRewards() external {
        Reporter storage reporter = reporters[msg.sender];
        uint256 amount = reporter.claimableRewards;
        if (amount == 0) {
            revert NothingToClaim();
        }

        reporter.claimableRewards = 0;
        (bool ok,) = msg.sender.call{value: amount}("");
        require(ok, "Transfer failed"); // TODO: check here if we need a custom error
        emit SlashedRewardsClaimed(msg.sender, amount);
    }
}
