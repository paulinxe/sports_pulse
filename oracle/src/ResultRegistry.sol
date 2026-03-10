// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @notice Interface for MatchRegistry to check if a match is registered.
 */
interface IMatchRegistry {
    function registeredMatches(bytes32 matchId) external view returns (bool);
}

/**
 * @notice Manages result submission windows and reporter submissions for matches.
 * @dev Only the owner can signal the end of a match to open a submission window.
 */
contract ResultRegistry is Ownable {
    enum MatchStatus {
        INACTIVE, // default — window not yet opened
        OPEN, // submission window is active
        FINALISED, // consensus reached, result stored
        DISPUTED, // disagreement on first attempt, retry pending
        UNRESOLVABLE // disagreement after retry, permanently closed
    }

    struct MatchResult {
        MatchStatus status;
        uint256 windowClosesAt; // block.timestamp + 1 hour
        uint8 attemptNumber; // 1 or 2 (max one retry)
        uint8 homeScore; // set at finalisation
        uint8 awayScore; // set at finalisation
        uint8 validReporterCount; // confidence level indicator
    }

    struct SubmittedScore {
        uint8 homeScore;
        uint8 awayScore;
        bool submitted; // double submission guard
    }

    /// @notice The MatchRegistry contract used to verify match existence.
    IMatchRegistry public immutable matchRegistry;

    /// @notice Duration of the submission window (1 hour).
    uint256 public constant SUBMISSION_WINDOW = 1 hours;

    /// @notice Maximum number of reporters allowed per match.
    uint8 public constant MAX_REPORTERS = 5;

    /// @notice Mapping from matchId to its result data.
    mapping(bytes32 => MatchResult) public results;

    /// @notice Mapping from matchId to reporter address to their submitted score.
    mapping(bytes32 => mapping(address => SubmittedScore)) public submissions;

    /// @notice List of reporters who submitted for each match (for iteration during consensus).
    mapping(bytes32 => address[]) public matchReporters;

    event SubmissionWindowOpened(bytes32 indexed matchId, uint256 windowClosesAt);

    error MatchNotRegistered(bytes32 matchId);
    error SubmissionWindowAlreadyOpened(bytes32 matchId);

    constructor(address _matchRegistry) Ownable(msg.sender) {
        matchRegistry = IMatchRegistry(_matchRegistry);
    }

    /**
     * @notice Signal that a match has ended and open the submission window.
     * @param matchId The unique identifier of the match.
     * @dev Only callable by the owner. Requires the match to be registered in MatchRegistry.
     */
    function signalMatchEnd(bytes32 matchId) external onlyOwner {
        if (!matchRegistry.registeredMatches(matchId)) {
            revert MatchNotRegistered(matchId);
        }

        if (results[matchId].status != MatchStatus.INACTIVE) {
            revert SubmissionWindowAlreadyOpened(matchId);
        }

        MatchResult storage result = results[matchId];
        result.status = MatchStatus.OPEN;
        result.windowClosesAt = block.timestamp + SUBMISSION_WINDOW;
        result.attemptNumber = 1;

        emit SubmissionWindowOpened(matchId, result.windowClosesAt);
    }
}
