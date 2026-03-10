// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

/**
 * @title ReporterRegistry
 * @notice Staking and Reporter eligibility. Reporters stake ETH to participate; slashing and rewards are applied here.
 */
contract ReporterRegistry {
    struct Reporter {
        uint256 stakedBalance; // current staked ETH
        uint256 claimableRewards; // accumulated slashed ETH from honest submissions
        uint256 withdrawalRequestedAt; // block.timestamp of withdrawal request, 0 if none
        uint256 correctSubmissions; // reputation tracking
        uint256 incorrectSubmissions; // reputation tracking
    }

    mapping(address => Reporter) public reporters;

    /// @notice Only this address may call slash().
    address public immutable consensusEngine;

    /// @notice Only this address may call cancelWithdrawalRequest().
    address public immutable resultRegistry;

    uint256 public constant MIN_STAKE = 0.1 ether;
    uint256 public constant WITHDRAWAL_COOLDOWN = 7 days;
    uint8 public constant MAX_REPORTERS = 5;
    uint8 public constant SLASH_PERCENTAGE = 25; // out of 100

    event Staked(address indexed reporter, uint256 amount);
    event WithdrawalRequested(address indexed reporter, uint256 claimableAt);
    event Withdrawn(address indexed reporter, uint256 amount);
    event Slashed(address indexed reporter, uint256 amount);
    event SlashedRewardsClaimed(address indexed reporter, uint256 amount);
    event WithdrawalRequestCancelled(address indexed reporter);

    error ZeroStake();
    error NoStakedBalance();
    error WithdrawalAlreadyRequested();
    error WithdrawalNotRequested();
    error CooldownNotElapsed(uint256 claimableAt);
    error NothingToClaim();
    error OnlyConsensusEngine();
    error OnlyResultRegistry();
    error ConsensusEngineZero();
    error ResultRegistryZero();
    error ExceedsMaxReporters();
    error ZeroReporterAddress();
    error ZeroCorrectReporters();

    modifier onlyConsensusEngine() {
        if (msg.sender != consensusEngine) {
            revert OnlyConsensusEngine();
        }
        _;
    }

    modifier onlyResultRegistry() {
        if (msg.sender != resultRegistry) {
            revert OnlyResultRegistry();
        }
        _;
    }

    constructor(address _consensusEngine, address _resultRegistry) {
        if (_consensusEngine == address(0)) {
            revert ConsensusEngineZero();
        }

        if (_resultRegistry == address(0)) {
            revert ResultRegistryZero();
        }

        consensusEngine = _consensusEngine;
        resultRegistry = _resultRegistry;
    }

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
     * @notice Cancel a pending withdrawal request. Only callable by ResultRegistry.
     * @param reporter The reporter whose withdrawal request should be cancelled.
     * @dev Called when a reporter submits a score while a withdrawal is pending.
     *      This ensures a reporter cannot be actively participating and exiting simultaneously.
     */
    function cancelWithdrawalRequest(address reporter) external onlyResultRegistry {
        Reporter storage r = reporters[reporter];
        if (r.withdrawalRequestedAt == 0) {
            return;
        }

        r.withdrawalRequestedAt = 0;
        emit WithdrawalRequestCancelled(reporter);
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

    /**
     * @notice Slash wrong reporters and distribute to correct reporters. Only callable by Consensus Engine.
     * @param wrongReporters Reporters who submitted an incorrect result; each loses 25% of staked balance.
     * @param correctReporters Reporters who submitted the correct result; share slashed ETH equally.
     */
    function slash(address[] calldata wrongReporters, address[] calldata correctReporters)
        external
        onlyConsensusEngine
    {
        if (wrongReporters.length + correctReporters.length > MAX_REPORTERS) {
            revert ExceedsMaxReporters();
        }

        uint256 correctReportersLength = correctReporters.length;
        if (correctReportersLength == 0) {
            // This should never happen. Therefore, revert.
            revert ZeroCorrectReporters();
        }

        uint256 totalSlashed = 0;
        uint256 wrongReportersLength = wrongReporters.length;
        for (uint256 i = 0; i < wrongReportersLength; i++) {
            Reporter storage reporter = reporters[wrongReporters[i]];
            uint256 stakedBalance = reporter.stakedBalance;
            if (stakedBalance == 0) {
                continue;
            }

            uint256 slashAmount = (stakedBalance * SLASH_PERCENTAGE) / 100;
            if (slashAmount > 0) {
                reporter.stakedBalance = stakedBalance - slashAmount;
                totalSlashed += slashAmount;
                emit Slashed(wrongReporters[i], slashAmount);
            }
        }

        if (totalSlashed == 0) {
            return;
        }

        uint256 rewardShare = totalSlashed / correctReportersLength;
        for (uint256 i = 0; i < correctReportersLength; i++) {
            reporters[correctReporters[i]].claimableRewards += rewardShare;
        }
    }
}
