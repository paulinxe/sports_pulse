// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

contract CompetitionRegistry is Ownable {

    uint32 public competitionIdCounter;
    mapping(uint32 => string) public competitions;

    event CompetitionAdded(uint32 indexed competitionId, string competitionName);

    constructor(string[] memory competitionNames) Ownable(msg.sender) {
        // TODO: revert if emptry string
        for (uint32 i = 0; i < competitionNames.length; i++) {
            competitionIdCounter++;
            competitions[competitionIdCounter] = competitionNames[i];
        }
    }

    function addCompetition(string memory competitionName) external onlyOwner {
        // TODO: revert if emptry string
        competitionIdCounter++;
        competitions[competitionIdCounter] = competitionName;
        emit CompetitionAdded(competitionIdCounter, competitionName);
    }
}