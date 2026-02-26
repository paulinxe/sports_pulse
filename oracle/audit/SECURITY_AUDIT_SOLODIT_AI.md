# Solodit-Based Security Audit — Oracle `/src` and Deployment (AI-Generated)

**Disclaimer:** This audit was generated using AI. It compares the SportsPulse oracle implementation to vulnerability patterns in the [Solodit](https://solodit.cyfrin.io) database. It is not a substitute for a professional security audit.

**Code version:** Commit `3406c52a9643dc6b05b4299ccb16f106ca9c83e5`

---

## Scope

- **Contracts:** `oracle/src/MatchRegistry.sol`, `CompetitionRegistry.sol`, `TeamRegistry.sol`
- **Deployment:** `oracle/script/Deploy.s.sol`
- **Patterns:** Oracle/registry design, EIP-712 signatures, signer rotation, access control, replay protection, deployment and ownership setup

---

## How This Audit Was Produced

1. **Searches performed:** Queries for `oracle`, `EIP712 signature replay`, `registry access control`, `authorized signer replay`, `domain separator EIP712`, and `oracle manipulation price` were run against the Solodit Findings API([https://solodit.cyfrin.io](https://solodit.cyfrin.io)).
2. **Comparison:** The returned findings were compared to the implementation in `oracle/src` and `oracle/script/Deploy.s.sol` to assess applicability and mitigation status.

---

## Implementation Summary


| Contract / Script       | Role                                                                     | Relevant security aspects                                      |
| ----------------------- | ------------------------------------------------------------------------ | -------------------------------------------------------------- |
| **MatchRegistry**       | Registers match results by validating EIP-712 signature from signer      | Signer validation, replay protection via `matchId`, registries |
| **CompetitionRegistry** | ID → competition name (owner-only writes)                                | Access control (Ownable), input validation                     |
| **TeamRegistry**        | ID → team name (owner-only writes)                                       | Access control (Ownable), input validation                     |
| **Deploy.s.sol**        | Deploys registries, sets signer, transfers ownership to `contractsOwner` | Input validation, role separation, ownership transfer          |


---

## Solodit Findings Used for Comparison

### 1. Signature replay & EIP-712

*Status: **Not vulnerable** = our implementation is not affected; **Not applicable** = finding does not apply to our design; **Mitigated** = we have protections in place; **Recommendation** = operational check, not a code bug.*


| Finding                                                                                                | Impact | Status             | Solodit link                                                                                                 | Relevance to this implementation                                                                                                                                                   |
| ------------------------------------------------------------------------------------------------------ | ------ | ------------------ | ------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [M-09] Possible scenario for Signature Replay Attack (no deadline, cross-chain, multi-contract replay) | MEDIUM | **Not vulnerable** | [Finding 8860](https://solodit.cyfrin.io/findings/8860)                                                      | MatchRegistry uses EIP-712 with `_hashTypedDataV4` (domain includes chainId and verifyingContract). Replay on same match is prevented by `matchId` and `MatchAlreadySubmitted`.    |
| Replay Attacks on Co-signer Signed Invocations                                                         | HIGH   | **Not vulnerable** | [Finding 19316](https://solodit.cyfrin.io/findings/19316)                                                    | One submission per `matchId`; signature bound to contract and chain via EIP-712 domain.                                                                                            |
| Signatures can be reused in perpetuity                                                                 | HIGH   | **Not vulnerable** | [Finding 63777](https://solodit.cyfrin.io/findings/63777)                                                    | Each `matchId` can be submitted only once; no reuse of the same signed payload.                                                                                                    |
| [M-04] verifyingContract set incorrectly for EIP712 Domain Separator                                   | MEDIUM | **Recommendation** | [Finding 18713](https://solodit.cyfrin.io/findings/18713)                                                    | We use OpenZeppelin EIP712("SportsPulse", "1"). Ensure the signer service uses the same `verifyingContract` (deployed MatchRegistry address) and chainId—operational/config check. |
| EIP712 DOMAIN_SEPARATOR stored as immutable                                                            | HIGH   | **Not vulnerable** | [Finding 27801](https://solodit.cyfrin.io/findings/27801)                                                    | OZ EIP712 uses a fixed domain at deployment; no upgrade path that could mismatch the separator.                                                                                    |
| Signed swap digest lacks a domain separator                                                            | MEDIUM | **Not vulnerable** | [Finding 63032](https://solodit.cyfrin.io/findings/63032)                                                    | Digest is built with `_hashTypedDataV4(structHash)`, which includes the domain separator.                                                                                          |
| [H-01] Signature replay in signatureClaim results in unauthorized claiming                             | HIGH   | **Not vulnerable** | [Finding 41087](https://solodit.cyfrin.io/findings/41087)                                                    | Match data is bound to `matchId` and one-time submission; no reward-claim style reuse.                                                                                             |
| LGO Vulnerable to Replay Attacks                                                                       | HIGH   | **Not vulnerable** | [Finding 60874](https://solodit.cyfrin.io/findings/60874)                                                    | EIP-712 domain binds signatures to chain and contract.                                                                                                                             |
| Cross-chain replay due to static DOMAIN_SEPARATOR                                                      | LOW    | **Not vulnerable** | [Finding 42685](https://solodit.cyfrin.io/findings/42685), [23721](https://solodit.cyfrin.io/findings/23721) | OZ EIP712 includes chainId in the domain.                                                                                                                                          |


### 2. Oracle / data consistency & manipulation

*Status: **Not vulnerable** = our implementation is not affected; **Not applicable** = finding does not apply to our design; **Mitigated** = we have protections in place; **Recommendation** = operational check, not a code bug.*


| Finding                                                          | Impact | Status             | Solodit link                                              | Relevance to this implementation                                                                                                   |
| ---------------------------------------------------------------- | ------ | ------------------ | --------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| M-9: Oracle Price miss matched when E-mode uses single oracle    | MEDIUM | **Not applicable** | [Finding 20223](https://solodit.cyfrin.io/findings/20223) | No price feed; single authorized signer and consistency derived from `matchId` (competitionId, homeTeamId, awayTeamId, matchDate). |
| [M-05] Front-running admin setPrice / single oracle manipulation | MEDIUM | **Not applicable** | [Finding 15988](https://solodit.cyfrin.io/findings/15988) | No on-chain “setPrice”; signer is off-chain. Owner can only rotate signer (old signers cannot be reused per `signersHistory`).     |
| [M-06] Uniswap oracle prices can be manipulated                  | MEDIUM | **Not applicable** | [Finding 45954](https://solodit.cyfrin.io/findings/45954) | No DEX or price oracle in scope.                                                                                                   |
| [H-04] Oracle price can be manipulated                           | HIGH   | **Not applicable** | [Finding 32056](https://solodit.cyfrin.io/findings/32056) | Match results are attestations from one authorized signer, not on-chain price feeds.                                               |
| Compromise of a single oracle enables limited price manipulation | HIGH   | **Recommendation** | [Finding 17709](https://solodit.cyfrin.io/findings/17709) | A compromised signer can submit wrong match results; mitigate with key security and signer rotation.                               |


### 3. Registry & access control

*Status: **Not vulnerable** = our implementation is not affected; **Not applicable** = finding does not apply to our design; **Mitigated** = we have protections in place; **Recommendation** = operational check, not a code bug.*


| Finding                                                                         | Impact | Status             | Solodit link                                              | Relevance to this implementation                                                                                         |
| ------------------------------------------------------------------------------- | ------ | ------------------ | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Potential for Disrupted Access Control Due to Registry Mismatch                 | LOW    | **Not applicable** | [Finding 59380](https://solodit.cyfrin.io/findings/59380) | CompetitionRegistry and TeamRegistry are immutable references set at MatchRegistry deployment; no runtime registry swap. |
| Lack of Access Control in Pong handlers                                         | HIGH   | **Not vulnerable** | [Finding 51879](https://solodit.cyfrin.io/findings/51879) | Critical actions are owner-only (CompetitionRegistry, TeamRegistry) or signature-gated (MatchRegistry).                  |
| Wrong permission control allows malicious Registry/Factory admin to steal funds | HIGH   | **Not applicable** | (Registry/Factory pattern)                                | Registries hold metadata only; MatchRegistry holds no user funds.                                                        |
| Missing zero address validation for authorized signer                           | LOW    | **Not vulnerable** | (Common pattern)                                          | Constructor and `rotateSigner` revert on `address(0)` with `InvalidAuthorizedSigner`.                                    |
| Signer reuse                                                                    | —      | **Not vulnerable** | —                                                         | `signersHistory` prevents reusing a previous signer address in `rotateSigner`.                                           |


### 4. Deployment script (Deploy.s.sol)

*Status: **Not vulnerable** = our implementation is not affected; **Not applicable** = finding does not apply to our design; **Mitigated** = we have protections in place; **Recommendation** = operational check, not a code bug.*

The script deploys `CompetitionRegistry`, `TeamRegistry`, and `MatchRegistry` in order, then transfers ownership of all three to `contractsOwner`. It reads competition and team names from JSON files (`script/data/competitions.json`, `script/data/teams.json`) and passes `authorizedSigner` and the two registries into `MatchRegistry`’s constructor.


| Concern                                            | Status             | Relevance to Deploy.s.sol                                                                                                                                               |
| -------------------------------------------------- | ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Zero address for signer or owner                   | **Not vulnerable** | Script `require`s `authorizedSigner != address(0)` and `contractsOwner != address(0)` before any deployment.                                                            |
| Deployer remains owner (ownership not transferred) | **Not vulnerable** | Ownership of all three contracts is explicitly transferred to `contractsOwner` via `transferOwnership(contractsOwner)`.                                                 |
| Single role (deployer = owner = signer)            | **Not vulnerable** | Script requires deployer, `authorizedSigner`, and `contractsOwner` to be distinct (`require(authorizedSigner != contractsOwner)`, deployer ≠ signer, deployer ≠ owner). |
| Non-atomic deployment                              | **Not applicable** | Script comments that deployment is not atomic; ownership can be transferred after some blocks. No funds at risk; registries hold only metadata.                         |
| Sensitive inputs from environment                  | **Recommendation** | `run()` reads `DEPLOYER_PRIVATE_KEY`, `AUTHORIZED_SIGNER_ADDRESS`, `CONTRACTS_OWNER_ADDRESS` from env. Use a secure env and avoid logging private keys.                 |


**Summary:** The deployment script enforces non-zero addresses and separation of deployer, signer, and owner, and transfers ownership to the intended owner. No Solodit finding links are cited for this section; the checks align with common deployment best practices (ownership transfer, role separation, input validation).

---

## Comparison Summary


| Area                                       | Implementation                                                    | Solodit pattern                                                | Status                       |
| ------------------------------------------ | ----------------------------------------------------------------- | -------------------------------------------------------------- | ---------------------------- |
| **Replay (same match)**                    | One submission per `matchId`; revert if already submitted         | Replay of signed message for same action                       | Mitigated                    |
| **Replay (cross-chain / other contracts)** | EIP-712 with domain (name, version, chainId, verifyingContract)   | Cross-chain or multi-contract replay when domain missing/wrong | Mitigated (OZ EIP712)        |
| **Signer rotation**                        | Owner can rotate; old signers cannot be reused (`signersHistory`) | Reuse of old signer / key compromise                           | Mitigated                    |
| **Zero signer**                            | Constructor and `rotateSigner` reject `address(0)`                | Missing zero check for signer                                  | Mitigated                    |
| **Registry admin**                         | Ownable; registries hold metadata only, no custody                | Malicious registry admin stealing funds                        | N/A (no funds in registries) |
| **Oracle/price manipulation**              | No price feed; signer attestations only                           | Price feed manipulation                                        | N/A                          |


---

## Recommendations

1. **EIP-712 domain:** Confirm that the signer service uses the same EIP-712 domain as the deployed MatchRegistry (same `verifyingContract` and chainId). Document this in runbooks.
2. **Owner and signer roles:** Keep owner (registry admin) and authorized signer as separate keys; limit owner key usage to configuration and signer rotation. The deployment script already enforces distinct deployer, signer, and owner.
3. **Operational security:** A compromised authorized signer can submit incorrect match results. Rely on key security, monitoring, and signer rotation; consider timelock or governance for signer rotation if the system grows.
4. **Deployment:** When running `Deploy.s.sol`, keep `DEPLOYER_PRIVATE_KEY`, `AUTHORIZED_SIGNER_ADDRESS`, and `CONTRACTS_OWNER_ADDRESS` in a secure environment (e.g. CI secrets or a secure env file); never log or commit private keys.

---

## Links to Solodit

- **Solodit:** [solodit.cyfrin.io](https://solodit.cyfrin.io)
- **Finding URLs:** Links use the form `https://solodit.cyfrin.io/findings/<id>`. If a link does not resolve, search by finding **ID** or **title** on Solodit.

---

*Audit generated with AI. Code version: commit 3406c52a9643dc6b05b4299ccb16f106ca9c83e5.*