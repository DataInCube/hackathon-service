# Hackathon Service Scope Boundary

This service is the hackathon orchestrator and system of record. It owns
hackathon lifecycle, rules, and submission intent, and emits events for other
services to react to. It does not implement team management, evaluation, or
leaderboards.

## In scope (owned here)
- Hackathon lifecycle and state machine (draft -> published -> warmup -> live -> submission_frozen -> evaluation_only -> completed -> archived)
- Tracks and versioned rules (immutable rule versions)
- Submission intent and submission state transitions (created, queued, invalidated)
- Team policy declaration and validation (requires/limits team size)
- Leaderboard policy declaration and freeze/publish requests (no ranking logic)
- Resources
- Data definitions (dataset metadata, files, variables, response schema)
- Evaluation metric definitions and submission limits
- Governance: reports, appeals, and audit log
- Domain events to NATS JetStream (hackathon.* and submission.*)

## Out of scope (owned by other services)
- Team creation, join/leave, membership management (team-service)
- Participant profiles and identity (auth/community services)
- Evaluation execution and scoring (evaluation-service)
- Ranking, sorting, and score storage (leaderboard-service)
- Chat/realtime collaboration (realtime/chat-service)

## Integration points
- Team service supplies `team_id` and membership; this service validates policy.
- Evaluation service consumes `submission.*` events and updates scores.
- Leaderboard service consumes freeze/publish events and exposes rankings.

## Events emitted
- `hackathon.created`, `hackathon.published`, `hackathon.phase.changed`
- `hackathon.team.required`, `hackathon.team.locked`, `hackathon.completed`
- `hackathon.rule.created`, `hackathon.rule.activated`
- `submission.created`, `submission.locked`, `submission.invalidated`
- `leaderboard.freeze.requested`, `leaderboard.publish.requested`
- `hackathon.data.*`, `hackathon.metric.*`, `hackathon.submission_limits.*`
