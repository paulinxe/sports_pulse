// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.30;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

contract TeamRegistry is Ownable {

    uint16 public teamIdCounter;
    mapping(uint16 => string) public teams;
    uint8 private constant MAX_TEAMS_PER_BATCH = 200;

    event TeamAdded(uint16 indexed teamId, string teamName);

    error TooManyTeams(uint256 numberOfTeams);
    error InvalidTeamName();

    constructor(string[] memory teamNames) Ownable(msg.sender) {
        revertIfTooManyTeams(teamNames.length);
        add(teamNames);
    }

    function addTeams(string[] memory teamNames) external onlyOwner {
        revertIfTooManyTeams(teamNames.length);
        add(teamNames);
    }

    function add(string[] memory teamNames) private {
        uint16 counter = teamIdCounter;

        for (uint16 i = 0; i < teamNames.length; i++) {
            revertIfEmptyString(teamNames[i]);

            counter++;
            teams[counter] = teamNames[i];
            emit TeamAdded(counter, teamNames[i]);
        }

        teamIdCounter = counter;
    }

    function revertIfTooManyTeams(uint256 numberOfTeams) private pure {
        if (numberOfTeams > MAX_TEAMS_PER_BATCH) {
            revert TooManyTeams(numberOfTeams);
        }
    }

    function revertIfEmptyString(string memory str) private pure {
        if (bytes(str).length == 0) {
            revert InvalidTeamName();
        }
    }
}
