// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

contract CompetitionRegistry is Ownable {

    uint32 public competitionIdCounter;
    mapping(uint32 => string) public competitions;
    uint8 private constant MAX_COMPETITIONS_PER_BATCH = 200;

    event CompetitionAdded(uint32 indexed competitionId, string competitionName);

    error TooManyCompetitions(uint8 numberOfCompetitions);

    constructor(string[] memory competitionNames) Ownable(msg.sender) {
        // TODO: revert if emptry string
        if (competitionNames.length > MAX_COMPETITIONS_PER_BATCH) {
            revert TooManyCompetitions(uint8(competitionNames.length));
        }

        for (uint32 i = 0; i < competitionNames.length; i++) {
            competitionIdCounter++;
            competitions[competitionIdCounter] = competitionNames[i];
        }
    }

    function addCompetition(string memory competitionName) external onlyOwner {
        // TODO: revert if emptry string
        // TODO: change this to allow a batch of teams to be added
        competitionIdCounter++;
        competitions[competitionIdCounter] = competitionName;
        emit CompetitionAdded(competitionIdCounter, competitionName);
    }
}