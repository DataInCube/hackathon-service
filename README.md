# hackathon-service

Hackathon orchestration and system-of-record service (Go + Postgres).  
Owns lifecycle, rules, tracks, submission intent, governance, and event emission.

## Changes & alignment with client spec
- Repositioned as orchestrator/system-of-record: removed team/participant ownership and leaderboard/evaluation computation.
- Enforced lifecycle state machine with explicit transitions (draft -> published -> warmup -> live -> submission_frozen -> evaluation_only -> completed -> archived).
- Added rule versioning with history, locking, and activation; submissions bind to locked rule versions.
- Added submission orchestration endpoints with evaluation status callbacks (no scoring logic here).
- Added team policy + validation endpoints (integration only).
- Added leaderboard integration hooks (freeze/unfreeze/publish) with events.
- Added governance (reports, appeals, audit logs).
- Added dataset metadata (including source URLs), files, variables, evaluation metrics, and submission limits.
- Emitted NATS JetStream events for all domain actions; added subjects/envs.
- Switched to UUID primary keys across models and schema.
- Improved validation and error responses for clearer client feedback.

See `docs/SCOPE_BOUNDARY.md` for scope boundaries and `docs/hackathon-service.postman_collection.json` for full Postman collection.

## Endpoints (base path: /api/v1)
Hackathons:
- POST /hackathons
- GET /hackathons
- GET /hackathons/{hackathonId}
- PUT /hackathons/{hackathonId}
- DELETE /hackathons/{hackathonId}
- POST /hackathons/{hackathonId}/publish
- POST /hackathons/{hackathonId}/transition
- GET /hackathons/{hackathonId}/state

Tracks & rules:
- POST /hackathons/{hackathonId}/tracks
- GET /hackathons/{hackathonId}/tracks
- GET /hackathons/{hackathonId}/tracks/{trackId}
- PUT /hackathons/{hackathonId}/tracks/{trackId}
- DELETE /hackathons/{hackathonId}/tracks/{trackId}
- GET /hackathons/{hackathonId}/rules
- POST /hackathons/{hackathonId}/rules
- GET /rules/{ruleId}
- PUT /rules/{ruleId}
- DELETE /rules/{ruleId}
- POST /rules/{ruleId}/version
- POST /rules/{ruleId}/versions
- POST /rules/versions/{ruleVersionId}/lock
- GET /rules/{ruleId}/history
- POST /hackathons/{hackathonId}/rules/{ruleVersionId}/activate

Rule versions are created in `draft` and must be locked before activation.

Team policy:
- GET /hackathons/{hackathonId}/team-policy
- POST /hackathons/{hackathonId}/teams/validate

Submissions:
- POST /hackathons/{hackathonId}/submissions
- GET /hackathons/{hackathonId}/submissions
- GET /submissions/{submissionId}
- PUT /submissions/{submissionId}
- DELETE /submissions/{submissionId}
- POST /submissions/{submissionId}/lock
- POST /submissions/{submissionId}/evaluation/start
- POST /submissions/{submissionId}/evaluation/fail
- POST /submissions/{submissionId}/evaluation/score
- POST /submissions/{submissionId}/invalidate

Leaderboard policy:
- GET /hackathons/{hackathonId}/leaderboard-policy
- POST /hackathons/{hackathonId}/leaderboard/freeze
- POST /hackathons/{hackathonId}/leaderboard/unfreeze
- POST /hackathons/{hackathonId}/leaderboard/publish

Resources & audit:
- GET /hackathons/{hackathonId}/resources
- POST /hackathons/{hackathonId}/resources
- GET /hackathons/{hackathonId}/resources/{resourceId}
- PUT /hackathons/{hackathonId}/resources/{resourceId}
- DELETE /hackathons/{hackathonId}/resources/{resourceId}
- POST /hackathons/{hackathonId}/reports
- POST /appeals
- GET /audit/hackathons/{hackathonId}

Data (datasets, files, variables):
- POST /hackathons/{hackathonId}/data
- GET /hackathons/{hackathonId}/data
- PUT /hackathons/{hackathonId}/data
- DELETE /hackathons/{hackathonId}/data
- POST /hackathons/{hackathonId}/data/files
- GET /hackathons/{hackathonId}/data/files
- GET /hackathons/{hackathonId}/data/files/{fileId}
- PUT /hackathons/{hackathonId}/data/files/{fileId}
- DELETE /hackathons/{hackathonId}/data/files/{fileId}
- POST /hackathons/{hackathonId}/data/variables
- GET /hackathons/{hackathonId}/data/variables
- GET /hackathons/{hackathonId}/data/variables/{variableId}
- PUT /hackathons/{hackathonId}/data/variables/{variableId}
- DELETE /hackathons/{hackathonId}/data/variables/{variableId}

Data notes:
- `source_urls` holds dataset/bucket links (e.g., GCS).
- `response_schema` lists target variables and submission format hints.
- Variables use `role` (feature/target/identifier) + optional `category`.

Evaluation metrics:
- POST /hackathons/{hackathonId}/metrics
- GET /hackathons/{hackathonId}/metrics
- GET /hackathons/{hackathonId}/metrics/{metricId}
- PUT /hackathons/{hackathonId}/metrics/{metricId}
- DELETE /hackathons/{hackathonId}/metrics/{metricId}

Metric notes:
- `scope`: `overall` or `per_target`.
- `target_variable` is required for `per_target`.

Submission limits:
- POST /hackathons/{hackathonId}/submission-limits
- GET /hackathons/{hackathonId}/submission-limits
- PUT /hackathons/{hackathonId}/submission-limits
- DELETE /hackathons/{hackathonId}/submission-limits

## Auth (Keycloak JWKS)
- AUTH_REQUIRED (default: true)
- AUTH_JWKS_URL (required when AUTH_REQUIRED=true)
- AUTH_ISSUER (Keycloak realm URL)
- AUTH_AUDIENCE (optional)
- AUTH_CLIENT_ID (client name for resource roles)

Role enforcement:
- hackathon_admin + hackathon_organizer: manage lifecycle, rules, tracks, leaderboard, audit
- other roles: can read hackathons and create submissions
- banned_user: blocked

## Events (NATS JetStream)
Events are published to JetStream for downstream services.

Subscribe
docker run --rm -it natsio/nats-box \
  nats --server nats://host.docker.internal:4222 sub "hackathon.>"

View stored events
docker run --rm -it natsio/nats-box \
  nats --server nats://host.docker.internal:4222 stream info SENTIO_EVENTS

docker run --rm -it natsio/nats-box \
  nats --server nats://host.docker.internal:4222 stream view SENTIO_EVENTS


Env:
- EVENTS_ENABLED (default: true)
- NATS_URL (default: nats://nats:4222)
- NATS_STREAM (default: SENTIO_EVENTS)
- NATS_SUBJECT_HACKATHON_CREATED (default: hackathon.created)
- NATS_SUBJECT_HACKATHON_PUBLISHED (default: hackathon.published)
- NATS_SUBJECT_HACKATHON_PHASE_CHANGED (default: hackathon.phase.changed)
- NATS_SUBJECT_HACKATHON_COMPLETED (default: hackathon.completed)
- NATS_SUBJECT_HACKATHON_DATA_CREATED (default: hackathon.data.created)
- NATS_SUBJECT_HACKATHON_DATA_UPDATED (default: hackathon.data.updated)
- NATS_SUBJECT_HACKATHON_DATA_DELETED (default: hackathon.data.deleted)
- NATS_SUBJECT_HACKATHON_DATA_FILE_CREATED (default: hackathon.data.file.created)
- NATS_SUBJECT_HACKATHON_DATA_FILE_UPDATED (default: hackathon.data.file.updated)
- NATS_SUBJECT_HACKATHON_DATA_FILE_DELETED (default: hackathon.data.file.deleted)
- NATS_SUBJECT_HACKATHON_DATA_VARIABLE_CREATED (default: hackathon.data.variable.created)
- NATS_SUBJECT_HACKATHON_DATA_VARIABLE_UPDATED (default: hackathon.data.variable.updated)
- NATS_SUBJECT_HACKATHON_DATA_VARIABLE_DELETED (default: hackathon.data.variable.deleted)
- NATS_SUBJECT_HACKATHON_METRIC_CREATED (default: hackathon.metric.created)
- NATS_SUBJECT_HACKATHON_METRIC_UPDATED (default: hackathon.metric.updated)
- NATS_SUBJECT_HACKATHON_METRIC_DELETED (default: hackathon.metric.deleted)
- NATS_SUBJECT_SUBMISSION_LIMITS_CREATED (default: hackathon.submission_limits.created)
- NATS_SUBJECT_SUBMISSION_LIMITS_UPDATED (default: hackathon.submission_limits.updated)
- NATS_SUBJECT_SUBMISSION_LIMITS_DELETED (default: hackathon.submission_limits.deleted)
- NATS_SUBJECT_SUBMISSION_CREATED (default: submission.created)
- NATS_SUBJECT_SUBMISSION_LOCKED (default: submission.locked)
- NATS_SUBJECT_SUBMISSION_INVALIDATED (default: submission.invalidated)
- NATS_SUBJECT_LEADERBOARD_FREEZE (default: leaderboard.freeze.requested)
- NATS_SUBJECT_LEADERBOARD_UNFREEZE (default: leaderboard.unfreeze.requested)
- NATS_SUBJECT_LEADERBOARD_PUBLISH (default: leaderboard.publish.requested)
- NATS_SUBJECT_TEAM_REQUIRED (default: hackathon.team.required)
- NATS_SUBJECT_TEAM_LOCKED (default: hackathon.team.locked)
- NATS_SUBJECT_RULE_CREATED (default: hackathon.rule.created)
- NATS_SUBJECT_RULE_VERSION_LOCKED (default: hackathon.rule.version.locked)
- NATS_SUBJECT_RULE_ACTIVATED (default: hackathon.rule.activated)

## Database
- Uses PostgreSQL with UUID primary keys. See schema: `schema.sql`.
- Connection pool:
  - DB_MAX_OPEN_CONNS (default: 10)
  - DB_MAX_IDLE_CONNS (default: 5)
  - DB_CONN_MAX_LIFETIME_MINUTES (default: 30)

## Run locally
```bash
go run ./cmd
```
