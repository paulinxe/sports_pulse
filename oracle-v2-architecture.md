# Football Oracle — V2 Architecture Plan

> **Status:** Planning / Pre-development  
> **Scope:** This document covers the product and architectural decisions for V2. No implementation code is included.

---

## Context & Goals

### V1 Baseline

The current V1 oracle has the following characteristics:

- Single permissioned signer (controlled by the protocol admin) — `authorizedSigner` in `MatchRegistry`
- EIP-712 signed score submissions — signer submits `(matchId, homeScore, awayScore)` via `submitMatch`
- Team list stored on-chain in `TeamRegistry`, seeded at deploy time (currently 20 LaLiga teams)
- Competition list stored on-chain in `CompetitionRegistry` (currently only LaLiga)
- One score per match — once submitted, `matches[matchId]` is immutable
- No reporter incentives or staking
- Match identity is `keccak256(abi.encodePacked(competitionId, homeTeamId, awayTeamId, matchDate))` where `matchDate` is `YYYYMMDD` (uint32)

### Why V2?

V1 is centralised in practice — a single signer means trust is entirely placed in one entity. V2 aims to introduce **genuine decentralisation** by allowing multiple independent reporters and economic accountability through staking and slashing. Reporter rewards and consumer fees are intentionally deferred to V3 — V2 focuses on getting the core protocol mechanics right first.

### Out of Scope for V2

- Commit-reveal scheme (deferred to V3)
- Token creation or token-based rewards (ETH only)
- Reporter rewards and consumer fees (deferred to V3)
- Dispute escalation to a DAO or human arbitration
- Off-chain access gating (accepted as a public good after finalization)

---

## Core Design Principles

1. **Permissionless reporting** — any address that meets the staking requirement can become a reporter
2. **ETH only** — no native token; staking and slashing are denominated in ETH
3. **No fees in V2** — consumers access results for free; reporters are not rewarded monetarily yet
4. **Slashing as deterrent** — reporters lose staked ETH for wrong submissions, incentivising honest behaviour even without rewards
5. **Historical data is a public good** — once a match is finalised, the result is free to read forever
6. **Admin controls scope, not outcomes** — the admin registers valid matches but has no influence over reported results

---

## Components

### 1. Team Registry

Already implemented as `TeamRegistry.sol`. The admin maintains a canonical on-chain registry of teams. Each team is identified by a `uint32` ID (auto-incremented, starting at 1). This prevents naming ambiguity (e.g. `"Man City"` vs `"Manchester City"`).

**Current state:** `teams.json` expanded with teams for all V2 competitions in scope.

**Responsibilities:**

- Store canonical team IDs and names
- Controlled by admin (transferable to DAO in future)
- Referenced by the Match Registry when registering matches
- `addTeams(string[] memory)` allows batch additions (capped at 200 per call)

---

### 2. Competition Registry

Already implemented as `CompetitionRegistry.sol`. The admin maintains a canonical on-chain registry of competitions. Each competition is identified by a `uint32` ID (auto-incremented, starting at 1).

**Current state:** `competitions.json` expanded with all V2 competitions in scope.

**Responsibilities:**

- Store canonical competition IDs and names
- Controlled by admin (transferable to DAO in future)
- Referenced by the Match Registry when registering matches
- `addCompetition(string memory)` adds one competition at a time (no batch in V1 — may be worth adding in V2)

---

### 3. Match Registry

In V1 this is `MatchRegistry.sol`, which conflates match registration and score submission into a single contract. In V2 the concerns will be separated: the Match Registry tracks upcoming matches, and the Result Registry handles submissions. The `MatchRegistry` name will be reused for this role.

**Match Identity:**
A match ID is a deterministic hash derived from:

- Competition ID (`uint32`)
- Season start year (`uint16` or `uint32`) — the calendar year when the season started (e.g. 2025 for 2025/26)
- Journey/phase number (`uint32`) — the matchday within the competition (e.g. Premier League has 38 journeys)
- Canonical home team ID (`uint32`)
- Canonical away team ID (`uint32`)

**Canonical form (human-readable):** `seasonYear_journey_homeTeamId_awayTeamId` — e.g. Liverpool (32) vs Newcastle (28) in journey 1 of the 2025/26 season → `2025_1_32_28`.

**Benefit:** If a match is rescheduled, it still belongs to the same journey; the identity is stable and date-agnostic.

**Caveat:** Without the season year, the same journey and team pair would collide when a new season starts. Prepending the season start year (e.g. `2025_1_32_28`) disambiguates seasons.

**On-chain:**

```solidity
matchId = keccak256(abi.encodePacked(competitionId, seasonYear, journey, homeTeamId, awayTeamId))
```

This replaces the V1 scheme that used match date (`YYYYMMDD`).

**Batch registration:** The admin registers matches via `registerBatch(MatchInput[] memory)`. Each element is a `MatchInput`:

```solidity
struct MatchInput {
    uint16 competitionId;
    uint16 seasonStartYear;
    uint8  journeyNumber;
    uint16 homeTeamId;
    uint16 awayTeamId;
}
```

- The array length is capped at **50** matches per call; if more are provided, the call reverts. This keeps gas consumption bounded.
- For each match in the batch, the contract emits a `**MatchRegistered`** event (one event per match).
- If any match in the batch is invalid (e.g. non-existing team or competition), the **entire transaction reverts** — no partial batches.

**Storage:** Matches are stored as `mapping(bytes32 => bool) public registeredMatches`. The key is the canonical match ID (the `bytes32` derived from the identity formula); the value is a boolean indicating that the match is registered.

**Responsibilities:**

- Maintain the list of valid upcoming matches
- Be the single source of truth for match existence
- Prevent consumers from requesting non-existent or duplicate matches

---

### 4. Reporter Registry

Replaces the V1 simple whitelist. Any address can become a reporter by meeting the minimum ETH stake requirement.

The Reporter Registry is also the staking contract — there is no need to separate them. Staking mechanics and reporter eligibility are tightly coupled: every staking action directly affects reporter status. Any address can become a reporter by staking at least the minimum required ETH.

**Responsibilities:**

*Staking:*
Data is stored in a single struct per reporter address:

```solidity
struct Reporter {
    uint256 stakedBalance;          // current staked ETH
    uint256 claimableRewards;       // accumulated slashed ETH from honest submissions
                                    // note: may be extracted to separate contract in V3
    uint256 withdrawalRequestedAt;  // block.timestamp of withdrawal request, 0 if none
    uint256 correctSubmissions;     // reputation tracking (used in V3)
    uint256 incorrectSubmissions;   // reputation tracking (used in V3)
}

mapping(address => Reporter) public reporters;
```

Functions:

- `stake()` — payable, accepts ETH and adds to `stakedBalance`. Serves both first-time staking and topping up after a slash. Emits `Staked`.
- `requestWithdrawal()` — sets `withdrawalRequestedAt` to current `block.timestamp`, starting the 7-day cooldown. If reporter submits a score while a withdrawal request is pending, `withdrawalRequestedAt` is automatically reset to 0 — a reporter cannot be actively participating and exiting simultaneously.
- `withdraw()` — claimable only after 7 days have passed since `withdrawalRequestedAt`. Transfers `stakedBalance` back to reporter and clears their entry. Emits `Withdrawn`.
- `claimSlashedRewards()` — transfers `claimableRewards` balance to reporter and resets it to 0. Pull model — never pushed automatically. Emits `SlashedRewardsClaimed`.

Events:

- `Staked(address indexed reporter, uint256 amount)`
- `WithdrawalRequested(address indexed reporter, uint256 claimableAt)`
- `Withdrawn(address indexed reporter, uint256 amount)`
- `Slashed(address indexed reporter, uint256 amount)`
- `SlashedRewardsClaimed(address indexed reporter, uint256 amount)`

*Eligibility:*

- `isEligible(address reporter) returns (bool)` — called by the Report Aggregator before accepting a submission
- A reporter is considered eligible if and only if their current staked balance is at or above the minimum stake threshold
- If a reporter is slashed below the minimum, they immediately lose eligibility and must top up before reporting again — no grace period
- Pending withdrawal doesn't affect eligibility
A reporter who has called `requestWithdrawal()` but is still within the 7-day cooldown still has their ETH locked and above the minimum. They are technically still isEligible = true during that window.

*Unstaking:*
Two-step process — request first, claim after cooldown:

```solidity
// Step 1 — start the 7-day cooldown
function requestWithdrawal() external

// Step 2 — claim ETH after cooldown has passed
function withdraw() external
```

`requestWithdrawal()` guards:

- Reporter must have a non-zero `stakedBalance`
- No pending withdrawal already in progress (`withdrawalRequestedAt == 0`)
- Sets `withdrawalRequestedAt = block.timestamp`

`withdraw()` guards:

- A withdrawal must have been requested (`withdrawalRequestedAt != 0`)
- 7 days must have passed: `block.timestamp >= withdrawalRequestedAt + 7 days`
- Transfers full `stakedBalance` to reporter
- Sets `stakedBalance` to 0 and resets `withdrawalRequestedAt` to 0 — the struct is **never deleted**, preserving `correctSubmissions` and `incorrectSubmissions` history for V3 reputation tracking

**Note:** `withdraw()` always exits the full balance. Partial withdrawals are not supported in V2 — a reporter is either fully in or fully out. This keeps the logic simple and avoids edge cases around partial slashing on a partially withdrawn balance.

*Slashing:*
Called by the Consensus Engine at finalization. Not callable by anyone else — the Reporter Registry must store the Consensus Engine contract address and restrict this function accordingly:

```solidity
function slash(
    address[] calldata wrongReporters,
    address[] calldata correctReporters
) external onlyConsensusEngine
```

Logic for each wrong reporter:

- Calculate slash amount: `slashAmount = stakedBalance * 25 / 100`
- Reduce `stakedBalance` by `slashAmount`
- Increment `incorrectSubmissions`

Logic for distributing to correct reporters:

- Sum total slashed ETH across all wrong reporters
- Divide equally: `rewardShare = totalSlashed / correctReporters.length`
- Add `rewardShare` to each correct reporter's `claimableRewards`
- Increment `correctSubmissions` for each correct reporter

**Note:** Integer division may leave a small remainder (e.g. 3 wrong reporters slashed, 2 correct reporters — division is not exact). The remainder stays in the contract. Handling of accumulated dust is deferred to V3.

**On Unbounded Array Protection:**
The `slash()` function accepts two arrays. To protect against unbounded array attacks, a hard limit is enforced using a protocol-level constant:

```solidity
uint8 public constant MAX_REPORTERS = 5;

require(wrongReporters.length + correctReporters.length <= MAX_REPORTERS, "Exceeds max reporters");
```

This constant is shared across the Reporter Registry and Consensus Engine — defined in one place (a shared interface or library) and referenced by both.

**On Sybil Attacks:**
A malicious actor could attempt to fill all MAX_REPORTERS slots with addresses they control, achieving majority vote and pushing a wrong result. The primary defence against this is **not the cap itself** but the staking requirement — controlling 3 out of 5 slots requires 0.3 ETH at risk, and a successful attack still triggers 25% slashing on each address (0.075 ETH lost). In V2 where there are no consumer fees or rewards, this attack is economically irrational — there is nothing to steal.

In V3 when real value flows through the system, both MAX_REPORTERS and the minimum stake must be recalibrated together against the actual value at risk per match. Increasing one without the other does not meaningfully improve security.

**V3: Timing Attack Mitigation**
A related attack is a malicious actor submitting results immediately when the submission window opens, filling slots before honest reporters can participate. V3 should consider enforcing a **minimum time between window opening and consensus running**, giving honest reporters a fair participation window before results are finalised.

**On Reputation:**
Reputation data is not used for any logic in V2 but is deliberately tracked from day one. In V3, well-behaved reporters (high accuracy, consistent participation) may be eligible for preferential rewards or airdrops. Building this history in V2 ensures that reporters who contribute early are recognised and rewarded retroactively in future versions.

**Parameters:**

- Minimum stake: 0.1 ETH (revisit in V3 when reporter rewards are introduced)
- Slash percentage: 25% per wrong submission (increase in V3)
- Unstaking cooldown: 7 days

---

### 5. Result Registry

The Result Registry owns the full submission lifecycle — from opening the submission window to storing the final consensus result. It is the contract consumers query for match results.

**On Match End Signalling:**
The blockchain cannot know when a real-world match finishes. Football matches vary in length due to injury time, extra time, and penalties. Two approaches exist:

- **V2 (Option A):** Admin calls `signalMatchEnd(matchId)` to open the submission window. Simple and reliable. This is a timing control only — the admin cannot influence what scores reporters submit.
- **V3 (Option B):** The submission window opens automatically at `kickoff + fixed buffer` (e.g. 3 hours to cover normal time + extra time + penalties), combined with the first reporter submission acting as an implicit signal. Removes admin dependency but requires careful buffer calibration.

**On EIP-712:**
V1 already uses EIP-712 (domain `"SportsPulse"`, version `"1"`) for the single-signer flow. V2 reuses the same signature scheme — reporters sign `(matchId, homeScore, awayScore)` using the same `MATCH_TYPEHASH`. The domain name and version are preserved so off-chain tooling needs minimal changes.

**Data Structures:**

```solidity
enum MatchStatus {
    INACTIVE,       // default — window not yet opened
    OPEN,           // submission window is active
    FINALISED,      // consensus reached, result stored
    DISPUTED,       // disagreement on first attempt, retry pending
    UNRESOLVABLE    // disagreement after retry, permanently closed
}

struct MatchResult {
    MatchStatus status;
    uint256 windowClosesAt;    // block.timestamp + 1 hour
    uint8 attemptNumber;       // 1 or 2 (max one retry)
    uint8 homeScore;           // set at finalisation
    uint8 awayScore;           // set at finalisation
    uint8 validReporterCount;  // confidence level indicator
}

struct SubmittedScore {
    uint8 homeScore;
    uint8 awayScore;
    bool submitted;            // double submission guard
}

mapping(bytes32 => MatchResult) public results;
mapping(bytes32 => mapping(address => SubmittedScore)) public submissions;
mapping(bytes32 => address[]) public matchReporters;  // iterable list for consensus
```

`**signalMatchEnd(bytes32 matchId)**` — `onlyOwner`

Guards:

- `matchRegistry.registeredMatches(matchId)` — match must exist (Match Registry exposes the mapping getter; no `isRegistered` helper in basecode)
- `results[matchId].status == INACTIVE` — cannot signal twice
- Sets `status = OPEN`, `windowClosesAt = block.timestamp + 1 hours`, `attemptNumber = 1`
- Emits `SubmissionWindowOpened(bytes32 indexed matchId, uint256 windowClosesAt)`

`**submitScore(bytes32 matchId, uint8 homeScore, uint8 awayScore, bytes calldata signature)**` — permissionless

Guards:

- `results[matchId].status == OPEN` — window must be open
- `block.timestamp <= results[matchId].windowClosesAt` — window must not have expired
- `reporterRegistry.isEligible(msg.sender)` — must be an eligible reporter (basecode exposes `isEligible`, not `isRegistered`)
- `!submissions[matchId][msg.sender].submitted` — no double submissions
- `matchReporters[matchId].length < MAX_REPORTERS` — reporter slots not full
- EIP-712 signature verification — recover signer from signature and verify it matches `msg.sender`

On success:

- Stores `SubmittedScore` for `msg.sender`
- Pushes `msg.sender` to `matchReporters[matchId]`
- Calls `reporterRegistry.cancelWithdrawalRequest(msg.sender)` — reporter cannot be exiting and participating simultaneously. Resets `withdrawalRequestedAt` to 0 when the reporter submits.
- Emits `ScoreSubmitted(bytes32 indexed matchId, address indexed reporter, uint8 homeScore, uint8 awayScore)`

#### 5.1 Consensus Engine

The consensus logic lives inside the Result Registry as a private function — not a separate contract. The algorithm is simple enough in V2 that a separate contract would introduce unnecessary struct duplication and cross-contract coupling. If V3 introduces weighted voting, commit-reveal, or reputation-based logic, extraction into a standalone contract becomes justified at that point.

**`finaliseMatch(bytes32 matchId)`** — permissionless

The public entry point that triggers consensus. Callable by anyone once the submission window has closed. Reporters are the most motivated callers as they want slashed rewards distributed as soon as possible.

Guards:
- `results[matchId].status == OPEN`
- `block.timestamp > results[matchId].windowClosesAt`

**Algorithm: Majority Vote + Confidence Level**
If a majority of submissions agree on the same `(homeScore, awayScore)` tuple, the match is finalised. The result is tagged with a confidence level based on the count of reporters who submitted the winning tuple. Any tie triggers the dispute flow.

**Logic:**
1. If `matchReporters[matchId].length == 0` → mark `DISPUTED` (no submissions)
2. Tally all submitted `(homeScore, awayScore)` tuples — find the most frequently submitted one:

```solidity
uint8 bestCount = 0;
uint8 bestHome = 0;
uint8 bestAway = 0;
bool tie = false;

for (uint8 i = 0; i < reporters.length; i++) {
    SubmittedScore memory s = submissions[matchId][reporters[i]];
    uint8 count = 0;

    for (uint8 j = 0; j < reporters.length; j++) {
        SubmittedScore memory t = submissions[matchId][reporters[j]];
        if (s.homeScore == t.homeScore && s.awayScore == t.awayScore) {
            count++;
        }
    }
    // note: i == j is intentional — every reporter counts themselves,
    // keeping all counts shifted by 1 equally. this avoids edge cases
    // where a sole unique submission would have count = 0.

    if (count > bestCount) {
        bestCount = count;
        bestHome = s.homeScore;
        bestAway = s.awayScore;
        tie = false;
    } else if (count == bestCount) {
        if (s.homeScore != bestHome || s.awayScore != bestAway) {
            tie = true; // two different tuples share the highest frequency
        }
    }
}
```

3. If `tie || bestCount == 0` → mark `DISPUTED`
4. If clear winner → mark `FINALISED`, store `bestHome` and `bestAway`, set `validReporterCount = bestCount`, assign confidence level
5. If `FINALISED` → build correct and wrong reporter arrays, then call `reporterRegistry.slash()`:

```solidity
address[] memory correctReporters = new address[](reporters.length);
address[] memory wrongReporters = new address[](reporters.length);
uint8 correctCount = 0;
uint8 wrongCount = 0;

for (uint8 i = 0; i < reporters.length; i++) {
    SubmittedScore memory s = submissions[matchId][reporters[i]];
    if (s.homeScore == bestHome && s.awayScore == bestAway) {
        correctReporters[correctCount] = reporters[i];
        correctCount++;
    } else {
        wrongReporters[wrongCount] = reporters[i];
        wrongCount++;
    }
}

// Trim arrays to their actual size before passing to slash().
// Both arrays were allocated with reporters.length slots upfront,
// so unfilled slots contain address(0). Without trimming, slash()
// would silently process those zero-padded slots — not reverting,
// but incorrectly incrementing correctSubmissions/incorrectSubmissions
// on address(0) and corrupting reputation data for a non-existent reporter.
assembly { mstore(correctReporters, correctCount) }
assembly { mstore(wrongReporters, wrongCount) }

reporterRegistry.slash(wrongReporters, correctReporters);
```

**Confidence Tiers:**

| Valid Reporters | Confidence Level | Recommended For |
| --------------- | ---------------- | ----------------------------------------------- |
| 1 | `VERY_LOW` | Informational use only |
| 2 | `LOW` | Low-stakes consumers |
| 3–4 | `MEDIUM` | General use |
| 5+ | `HIGH` | High-stakes consumers (e.g. prediction markets) |

These thresholds are protocol parameters and can be adjusted by the admin over time.

**What Consumers Receive:**
Every finalised result exposes two fields:
- **Result** — the agreed `(homeScore, awayScore)` tuple
- **`validReporterCount`** — the number of reporters that submitted this result

**Dispute Trigger:**
Confidence level does not affect whether a dispute is triggered. The dispute flow is triggered exclusively by reporter **disagreement**. Low participation with full agreement always finalises.

**Possible Outcomes:**

| Outcome | Condition | Next Step |
| -------------- | ------------------------------------ | ---------------------------------------------------------- |
| `FINALISED` | All submissions agree (>=1 reporter) | Confidence level assigned, slashing applied, result public |
| `DISPUTED` | Reporters disagree on first attempt | Retry mechanism activated |
| `UNRESOLVABLE` | Reporters disagree on retry attempt | Match permanently closed |

**V3:** Extract into a standalone `ConsensusEngine` contract once the algorithm grows in complexity (weighted voting, commit-reveal, reputation weighting).

---

### 6. Dispute & Retry Mechanism

If consensus is not reached, the oracle enters a retry flow without any manual intervention.

**Flow:**

1. Match marked as `DISPUTED`
2. A cooldown period begins automatically (duration TBD — does not affect architecture)
3. After the cooldown, the submission window reopens and reporters may resubmit
4. Consensus Engine runs again
5. If all resubmissions agree → `FINALISED` (with confidence level based on count)
6. If reporters still disagree → `UNRESOLVABLE`

**There is exactly one retry.** A match that reaches `UNRESOLVABLE` is permanently closed.

---

### 7. Slashing

Triggered at finalization. Wrong reporters lose 25% of their staked ETH. The slashed amount is distributed equally among the honest reporters (those who submitted the winning result) as a claimable reward — using a pull model, never pushed automatically.

**Wrong reporters:**

- Lose 25% of their staked ETH per wrong submission
- This is intentionally moderate in V2 since reporters earn no rewards yet — harsh slashing with no upside would deter participation

**Honest reporters:**

- Receive an equal share of the total slashed ETH from that match
- Amount becomes claimable immediately after finalization
- The more reporters that lied, the larger the honest reporters' share — the system naturally rewards reporters more when attacks occur

**Example:**

> 5 reporters submit. 4 agree on 2-1, 1 submits 3-0. Minimum stake is 0.1 ETH, slash is 25% = 0.025 ETH slashed. Each of the 4 honest reporters can claim 0.00625 ETH.

**V3 Improvements:**

- Slash percentage to be increased once reporters are earning rewards — asymmetric risk/reward becomes viable
- Slashed ETH distribution weighted by reporter reputation rather than equal split — long-standing honest reporters earn a larger share
- Commit-reveal will help distinguish genuine data source errors from malicious submissions, enabling more nuanced slashing logic

---

## Match Lifecycle

```
Admin registers match in Match Registry
            ↓
Match is played
            ↓
Admin signals match end
            ↓
Reporter submission window opens (~1 hour)
            ↓
Reporters submit EIP-712 signed scores
            ↓
Submission window closes
            ↓
Consensus Engine runs
            ↓
   ┌─────────┴─────────┐
All agree (≥1)    Disagreement
   ↓                    ↓
FINALISED           DISPUTED
   ↓                    ↓
Confidence level    Cooldown period
assigned (VERY_LOW      ↓
→ LOW → MEDIUM →   Resubmission window
HIGH by count)          ↓
   ↓            Consensus Engine runs
Slash wrong             ↓
reporters 25%    ┌──────┴──────┐
   ↓         All agree    Disagreement
Slashed ETH      ↓              ↓
split equally FINALISED   UNRESOLVABLE
among honest
reporters
(claimable)
Result public forever
```

---

## Key Decisions Log


| Topic                   | Decision                                                                | Notes                                                                                       |
| ----------------------- | ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| Currency                | ETH only                                                                | No native token                                                                             |
| Staking asset           | ETH                                                                     | Reporters stake ETH                                                                         |
| Minimum stake           | 0.1 ETH                                                                 | Meaningful commitment without blocking participation; revisit in V3                         |
| Slashing asset          | ETH                                                                     | Wrong reporters lose staked ETH                                                             |
| Slash percentage        | 25% per wrong submission                                                | Moderate in V2 since reporters earn no rewards yet; increase in V3                          |
| Slashed ETH destination | Equal split among honest reporters                                      | Claimable via pull model; weighted by reputation in V3                                      |
| Reporter rewards        | Slashed ETH only in V2                                                  | Consumer fee rewards deferred to V3                                                         |
| Consumer fees           | None in V2                                                              | Deferred to V3                                                                              |
| Consumer access         | Always free                                                             | Result is public after finalization                                                         |
| Historical data         | Free after finalization                                                 | Accepted as public good                                                                     |
| Consensus algorithm     | Unanimous agreement                                                     | All submissions must agree; any disagreement triggers dispute                               |
| Quorum                  | Replaced by confidence model                                            | Hard quorum dropped; see confidence tiers                                                   |
| Confidence model        | VERY_LOW / LOW / MEDIUM / HIGH                                          | Based on valid reporter count; consumers choose their own threshold                         |
| Confidence thresholds   | 1 / 2 / 3–4 / 5+ reporters                                              | Protocol parameters, adjustable by admin                                                    |
| Dispute trigger         | Reporter disagreement only                                              | Low participation with full agreement always finalises                                      |
| Commit-reveal           | Deferred to V3                                                          | Not needed for football scores in V2                                                        |
| Dispute retry           | 1 automatic retry after cooldown                                        | No manual intervention                                                                      |
| Unresolvable matches    | Permanently closed                                                      | No fee pool to worry about in V2                                                            |
| Withdrawal delay        | Yes (duration TBD)                                                      | Prevents reporters escaping slashing                                                        |
| Match end signalling    | Admin-controlled in V2                                                  | Timing only, not result control                                                             |
| Reporter reputation     | Tracked from V2, unused until V3                                        | Enables retroactive rewards and airdrops                                                    |
| Match ID scheme         | `keccak256(competitionId, seasonYear, journey, homeTeamId, awayTeamId)` | V2: journey/phase + season year (reschedule-stable; no date); replaces V1 date-based scheme |
| Team/competition IDs    | `uint32`, auto-incremented from 1                                       | Preserves V1 type; `uint16` would be too narrow for multi-league growth                     |


---

## Open Items for Future Iterations

- Consider adding `addCompetitions(string[] memory)` batch function to `CompetitionRegistry` for consistency with `TeamRegistry`
- Cooldown period duration for disputed matches
- Withdrawal delay duration
- Confidence tier thresholds to be reviewed once real reporter pool size is known (currently: 1 / 2 / 3–4 / 5+)
- **V3: Increase slash percentage now that reporters earn rewards**
- **V3: Weight slashed ETH distribution by reporter reputation instead of equal split**
- **V3: Consumer fee model and reporter rewards from fees**
- **V3: Airdrop or preferential reward logic for well-behaved reporters based on V2 reputation history**
- **V3: Automatic match end signalling (kickoff buffer + first submission)**
- **V3: Commit-reveal to distinguish honest data errors from malicious submissions**
- **V3: Confidence weighting by reporter reputation, not just count** — a result from 2 high-reputation reporters may be more trustworthy than 5 new ones
- **V3: Consumer-side on-chain confidence threshold enforcement** — consumers register their minimum accepted confidence level; oracle only delivers results that meet it
- V3+: DAO governance for admin functions

