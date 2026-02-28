# Sports Pulse

Push Oracle that fetches football (at least for now) match results, signs them with EIP-712, and submits them to an on-chain Oracle.

**Learning project** — This repo is for learning purposes; there are no plans to deploy to a live blockchain in the near future.

Sports Pulse brings football match results from external data providers onto the blockchain. Raw match data is synced by the provider, cryptographically signed by the signer, relayed to the chain by the relayer, and finally validated and stored by the Oracle smart contract.

## Components

For a detailed description, please access the README.md of the respective service.

- **Provider** — Fetches match data from third-party providers and syncs it into the system (only matches that have finished).
- **Signer** — Signs each match with EIP-712 (matchId, homeScore, awayScore) for on-chain verification.
- **Relayer** — Picks up signed matches and submits them to the Oracle smart contract via transactions.
- **Oracle** — On-chain contract that validates signatures and stores match results.

## Deployment (Sepolia)

Contract addresses from the Sepolia deployment:

| Contract             | Address |
|----------------------|---------|
| CompetitionRegistry  | `0x1a2429256539DA04887Ac0CD9c38327f96b81BC7` |
| TeamRegistry         | `0xBeCA2f4BaE3be06D73c7a028bde5647783f8956c` |
| MatchRegistry        | `0x4C46121Db184A78f8bDb2bbadf68b4e71101052d` |
| Authorized Signer    | `0x7869F2E6182C83F74A7054Ed29Bb29DE667Ee648` |
| Contracts Owner      | `0xc4FcA5D912E8841D01336Ad32d623814c0b45791` |

### First match broadcast

- **Match:** Alavés – Girona (23/02/2026)
- **Transaction:** [`0xe85d708276a1d1b2ac8ed4a15c1805d7d7f5ab0fc54961d91dbdc0a154c23215`](https://sepolia.etherscan.io/tx/0xe85d708276a1d1b2ac8ed4a15c1805d7d7f5ab0fc54961d91dbdc0a154c23215)
- **Canonical ID:** `0xb5eb022f67c0de11b4ee7885df10476f162ae0f5d2cf262e591d5d61c38bbf1b`

## Next steps

Post-MVP improvements:

- Accept more than 1 result on chain for the same match (N of M agreement)
- Target multiple chains
- Deploy off-chain services to the cloud
- Accept more sports
