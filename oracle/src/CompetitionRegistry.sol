// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

contract CompetitionRegistry is Ownable {

    uint8 public competitionIdCounter;
    mapping(uint8 => string) public competitions;
    uint8 private constant MAX_COMPETITIONS_PER_BATCH = 10;

    event CompetitionAdded(uint8 indexed competitionId, string competitionName);

    error TooManyCompetitions(uint256 numberOfCompetitions);
    error InvalidCompetitionName();

    constructor(string[] memory competitionNames) Ownable(msg.sender) {
        if (competitionNames.length > MAX_COMPETITIONS_PER_BATCH) {
            revert TooManyCompetitions(uint8(competitionNames.length));
        }

        uint8 counter = competitionIdCounter;

        for (uint8 i = 0; i < competitionNames.length; i++) {
            revertIfEmptyString(competitionNames[i]);
            
            counter++;
            competitions[counter] = competitionNames[i];
            emit CompetitionAdded(counter, competitionNames[i]);
        }

        competitionIdCounter = counter;
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
