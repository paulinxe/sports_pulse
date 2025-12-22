// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

contract CompetitionRegistry is Ownable {

    uint32 public competitionIdCounter;
    mapping(uint32 => string) public competitions;
    uint8 private constant MAX_COMPETITIONS_PER_BATCH = 200;

    event CompetitionAdded(uint32 indexed competitionId, string competitionName);

    error TooManyCompetitions(uint8 numberOfCompetitions);
    error InvalidCompetitionName();

    constructor(string[] memory competitionNames) Ownable(msg.sender) {
        if (competitionNames.length > MAX_COMPETITIONS_PER_BATCH) {
            revert TooManyCompetitions(uint8(competitionNames.length));
        }

        for (uint32 i = 0; i < competitionNames.length; i++) {
            revertIfEmptyString(competitionNames[i]);

            competitionIdCounter++;
            competitions[competitionIdCounter] = competitionNames[i];
        }
    }

    function addCompetition(string memory competitionName) external onlyOwner {
        revertIfEmptyString(competitionName);
        
        competitionIdCounter++;
        competitions[competitionIdCounter] = competitionName;
        emit CompetitionAdded(competitionIdCounter, competitionName);
    }

    function revertIfEmptyString(string memory str) private pure {
        if (bytes(str).length == 0) {
            revert InvalidCompetitionName();
        }
    }
}