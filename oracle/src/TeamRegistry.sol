// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

contract TeamRegistry is Ownable {

    uint32 public teamIdCounter;
    mapping(uint32 => string) public teams;
    uint8 private constant MAX_TEAMS_PER_BATCH = 200;

    event TeamAdded(uint32 indexed teamId, string teamName);

    error TooManyTeams(uint8 numberOfTeams);

    constructor(string[] memory teamNames) Ownable(msg.sender) {
        // TODO: revert if emptry string
        if (teamNames.length > MAX_TEAMS_PER_BATCH) {
            revert TooManyTeams(uint8(teamNames.length));
        }

        for (uint32 i = 0; i < teamNames.length; i++) {
            teamIdCounter++;
            teams[teamIdCounter] = teamNames[i];
        }
    }

    function addTeam(string memory teamName) external onlyOwner {
        // TODO: revert if emptry string
        // TODO: change this to allow a batch of teams to be added
        teamIdCounter++;
        teams[teamIdCounter] = teamName;
        emit TeamAdded(teamIdCounter, teamName);
    }
}