// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

contract TeamRegistry is Ownable {

    uint32 public teamIdCounter;
    mapping(uint32 => string) public teams;

    event TeamAdded(uint32 indexed teamId, string teamName);

    constructor(string[] memory teamNames) Ownable(msg.sender) {
        // TODO: revert if emptry string
        for (uint32 i = 0; i < teamNames.length; i++) {
            teamIdCounter++;
            teams[teamIdCounter] = teamNames[i];
        }
    }

    function addTeam(string memory teamName) external onlyOwner {
        // TODO: revert if emptry string
        teamIdCounter++;
        teams[teamIdCounter] = teamName;
        emit TeamAdded(teamIdCounter, teamName);
    }
}