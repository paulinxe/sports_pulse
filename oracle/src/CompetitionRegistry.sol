// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

contract CompetitionRegistry is Ownable {

    uint32 public competitionIdCounter;
    mapping(uint32 => string) public competitions;

    // TODO: here we need to allow to accept a set of predefined leagues
    constructor() Ownable(msg.sender) {}

    function addCompetition(string memory competitionName) external onlyOwner {
        competitionIdCounter++;
        competitions[competitionIdCounter] = competitionName;
    }
}