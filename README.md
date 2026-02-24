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
