# Milestone 2 Alignment (Hackathon Service)

This doc summarizes what changed in the hackathon-service for Milestone 2 and how it aligns with the client requirements.

## Scope alignment (client requirements -> implementation)

1) Service role and boundaries
- Implemented as orchestrator/system-of-record for hackathons, rules, tracks, submission intent, governance, and events.
- Explicitly excludes team creation/matching, leaderboard computation, evaluation logic, realtime/chat, and auth identity.
- Team, leaderboard, and evaluation remain separate services; this service only provides integration hooks.

2) Lifecycle as state machine
- States: draft, published, warmup, live, submission_frozen, evaluation_only, completed, archived.
- Transitions enforced by service logic (invalid transitions return clear messages).
- Live requires an active, locked rule version.
- Deletion only allowed in draft to protect audit history.

3) Rules and versioning
- Rules are versioned with history; versions are created as draft.
- Versions must be locked before activation.
- Submissions bind to the active rule version at time of submission.

4) Submission orchestration (metadata only)
- Stores intent + metadata; no evaluation or scoring logic.
- Evaluation-service can update status through dedicated callbacks.
- Submission lifecycle: created -> queued_for_evaluation -> evaluation_running -> evaluation_failed | scored | invalidated.

5) Team and matching integration
- Team policy exposed and validation endpoint for eligibility.
- Emits team required/locked events (no team CRUD in this service).

6) Leaderboard integration
- Exposes policy and freeze/unfreeze/publish triggers.
- Emits leaderboard integration events (no ranking or score computation).

7) Governance and audit trail
- Reports, appeals, and audit logs persisted.
- Audit logs generated for all key state changes and orchestration actions.

8) Events (NATS JetStream)
- Emits domain events for downstream services to react to.
- Subjects configurable via env; JetStream stream defaults to SENTIO_EVENTS.

9) Identity and authorization
- Keycloak JWT validation via JWKS (no identity ownership).
- Role-based enforcement for admin/organizer operations and evaluation callbacks.

## Endpoint coverage (methods + roles)

Base path: /api/v1

Hackathons
- POST /hackathons (admin/organizer)
- GET /hackathons (authenticated)
- GET /hackathons/{hackathonId} (authenticated)
- PUT /hackathons/{hackathonId} (admin/organizer)
- DELETE /hackathons/{hackathonId} (admin/organizer, only draft)
- POST /hackathons/{hackathonId}/publish (admin/organizer)
- POST /hackathons/{hackathonId}/transition (admin/organizer)
- GET /hackathons/{hackathonId}/state (authenticated)

Tracks
- POST /hackathons/{hackathonId}/tracks (admin/organizer)
- GET /hackathons/{hackathonId}/tracks (authenticated)
- GET /hackathons/{hackathonId}/tracks/{trackId} (authenticated)
- PUT /hackathons/{hackathonId}/tracks/{trackId} (admin/organizer)
- DELETE /hackathons/{hackathonId}/tracks/{trackId} (admin/organizer)

Rules
- POST /hackathons/{hackathonId}/rules (admin/organizer)
- GET /hackathons/{hackathonId}/rules (authenticated)
- GET /rules/{ruleId} (authenticated)
- PUT /rules/{ruleId} (admin/organizer)
- DELETE /rules/{ruleId} (admin/organizer)
- POST /rules/{ruleId}/version (admin/organizer)
- POST /rules/{ruleId}/versions (admin/organizer)
- POST /rules/versions/{ruleVersionId}/lock (admin/organizer)
- GET /rules/{ruleId}/history (authenticated)
- POST /hackathons/{hackathonId}/rules/{ruleVersionId}/activate (admin/organizer)

Team integration
- GET /hackathons/{hackathonId}/team-policy (authenticated)
- POST /hackathons/{hackathonId}/teams/validate (authenticated)

Submissions (metadata only)
- POST /hackathons/{hackathonId}/submissions (participant token)
- GET /hackathons/{hackathonId}/submissions (authenticated)
- GET /submissions/{submissionId} (authenticated)
- PUT /submissions/{submissionId} (owner or admin/organizer)
- DELETE /submissions/{submissionId} (owner or admin/organizer)
- POST /submissions/{submissionId}/lock (admin/organizer)
- POST /submissions/{submissionId}/invalidate (admin/organizer)

Evaluation callbacks (from evaluation-service)
- POST /submissions/{submissionId}/evaluation/start (admin/organizer/platform_admin/evaluation_executor)
- POST /submissions/{submissionId}/evaluation/fail (admin/organizer/platform_admin/evaluation_executor)
- POST /submissions/{submissionId}/evaluation/score (admin/organizer/platform_admin/evaluation_executor)

Leaderboard integration
- GET /hackathons/{hackathonId}/leaderboard-policy (authenticated)
- POST /hackathons/{hackathonId}/leaderboard/freeze (admin/organizer)
- POST /hackathons/{hackathonId}/leaderboard/unfreeze (admin/organizer)
- POST /hackathons/{hackathonId}/leaderboard/publish (admin/organizer)

Resources and governance
- GET /hackathons/{hackathonId}/resources (authenticated)
- POST /hackathons/{hackathonId}/resources (admin/organizer)
- GET /hackathons/{hackathonId}/resources/{resourceId} (authenticated)
- PUT /hackathons/{hackathonId}/resources/{resourceId} (admin/organizer)
- DELETE /hackathons/{hackathonId}/resources/{resourceId} (admin/organizer)
- POST /hackathons/{hackathonId}/reports (authenticated)
- POST /appeals (authenticated)
- GET /audit/hackathons/{hackathonId} (admin/organizer)

## Events emitted (contract)
- hackathon.created
- hackathon.published
- hackathon.phase.changed
- hackathon.completed
- hackathon.team.required
- hackathon.team.locked
- hackathon.rule.created
- hackathon.rule.version.locked
- hackathon.rule.activated
- submission.created
- submission.locked
- submission.invalidated
- leaderboard.freeze.requested
- leaderboard.unfreeze.requested
- leaderboard.publish.requested

## Data model changes
- UUID primary keys for all core entities.
- Rule versions include status (draft/locked) and lock timestamp.
- Audit logs persist action history for lifecycle and governance.

## Out of scope (by design)
- Team creation/joining/matching (handled by team-service).
- Leaderboard computation/storage (handled by leaderboard-service).
- Evaluation logic (handled by evaluation-service).
- Auth identity (handled by Keycloak/Sentio).

## Supporting files
- Scope boundaries: docs/SCOPE_BOUNDARY.md
- Postman collection: docs/hackathon-service.postman_collection.json
