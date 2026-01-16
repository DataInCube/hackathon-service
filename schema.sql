CREATE TABLE hackathons (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    state TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'public',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    allows_teams BOOLEAN NOT NULL DEFAULT false,
    requires_teams BOOLEAN NOT NULL DEFAULT false,
    min_team_size INTEGER NOT NULL DEFAULT 1,
    max_team_size INTEGER NOT NULL DEFAULT 0,
    active_rule_version_id UUID,
    leaderboard_frozen BOOLEAN NOT NULL DEFAULT false,
    leaderboard_published BOOLEAN NOT NULL DEFAULT false,
    created_by TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    archived_at TIMESTAMPTZ,
    CHECK (requires_teams = false OR allows_teams = true),
    CHECK (min_team_size >= 0),
    CHECK (max_team_size >= 0),
    CHECK (max_team_size = 0 OR min_team_size <= max_team_size)
);

CREATE INDEX hackathons_state_idx ON hackathons (state);

CREATE TABLE hackathon_datasets (
    id UUID PRIMARY KEY,
    hackathon_id UUID NOT NULL UNIQUE REFERENCES hackathons(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    source_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
    response_schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX hackathon_datasets_hackathon_id_idx ON hackathon_datasets (hackathon_id);

CREATE TABLE dataset_files (
    id UUID PRIMARY KEY,
    dataset_id UUID NOT NULL REFERENCES hackathon_datasets(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    file_type TEXT NOT NULL,
    description TEXT,
    url TEXT NOT NULL,
    size_bytes BIGINT,
    checksum TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (dataset_id, name)
);

CREATE INDEX dataset_files_dataset_id_idx ON dataset_files (dataset_id);

CREATE TABLE dataset_variables (
    id UUID PRIMARY KEY,
    dataset_id UUID NOT NULL REFERENCES hackathon_datasets(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    data_type TEXT NOT NULL,
    description TEXT,
    unit TEXT,
    category TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (dataset_id, name)
);

CREATE INDEX dataset_variables_dataset_id_idx ON dataset_variables (dataset_id);

CREATE TABLE evaluation_metrics (
    id UUID PRIMARY KEY,
    hackathon_id UUID NOT NULL REFERENCES hackathons(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    metric_type TEXT NOT NULL,
    direction TEXT NOT NULL,
    scope TEXT NOT NULL DEFAULT 'overall',
    target_variable TEXT,
    weight DOUBLE PRECISION NOT NULL DEFAULT 1,
    description TEXT,
    params JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (hackathon_id, name)
);

CREATE INDEX evaluation_metrics_hackathon_id_idx ON evaluation_metrics (hackathon_id);

CREATE TABLE submission_limits (
    id UUID PRIMARY KEY,
    hackathon_id UUID NOT NULL UNIQUE REFERENCES hackathons(id) ON DELETE CASCADE,
    per_day INTEGER NOT NULL DEFAULT 0,
    total INTEGER NOT NULL DEFAULT 0,
    per_team INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX submission_limits_hackathon_id_idx ON submission_limits (hackathon_id);

CREATE TABLE tracks (
    id UUID PRIMARY KEY,
    hackathon_id UUID NOT NULL REFERENCES hackathons(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (hackathon_id, name)
);

CREATE INDEX tracks_hackathon_id_idx ON tracks (hackathon_id);

CREATE TABLE rules (
    id UUID PRIMARY KEY,
    hackathon_id UUID NOT NULL REFERENCES hackathons(id) ON DELETE CASCADE,
    track_id UUID REFERENCES tracks(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX rules_hackathon_id_idx ON rules (hackathon_id);
CREATE INDEX rules_track_id_idx ON rules (track_id);

CREATE TABLE rule_versions (
    id UUID PRIMARY KEY,
    rule_id UUID NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    status TEXT NOT NULL,
    content JSONB NOT NULL,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    locked_at TIMESTAMPTZ,
    CHECK (status IN ('draft','locked')),
    UNIQUE (rule_id, version)
);

CREATE INDEX rule_versions_rule_id_idx ON rule_versions (rule_id);

CREATE TABLE submissions (
    id UUID PRIMARY KEY,
    hackathon_id UUID NOT NULL REFERENCES hackathons(id) ON DELETE CASCADE,
    track_id UUID REFERENCES tracks(id) ON DELETE SET NULL,
    rule_version_id UUID NOT NULL REFERENCES rule_versions(id) ON DELETE RESTRICT,
    submitted_by TEXT NOT NULL,
    team_id TEXT,
    status TEXT NOT NULL,
    phase TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    locked_at TIMESTAMPTZ,
    invalidated_at TIMESTAMPTZ
);

CREATE INDEX submissions_hackathon_id_idx ON submissions (hackathon_id);
CREATE INDEX submissions_status_idx ON submissions (status);
CREATE INDEX submissions_rule_version_idx ON submissions (rule_version_id);

CREATE TABLE resources (
    id UUID PRIMARY KEY,
    hackathon_id UUID NOT NULL REFERENCES hackathons(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX resources_hackathon_id_idx ON resources (hackathon_id);

CREATE TABLE reports (
    id UUID PRIMARY KEY,
    hackathon_id UUID NOT NULL REFERENCES hackathons(id) ON DELETE CASCADE,
    reporter_id TEXT,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX reports_hackathon_id_idx ON reports (hackathon_id);

CREATE TABLE appeals (
    id UUID PRIMARY KEY,
    submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    appellant_id TEXT,
    content TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX appeals_submission_id_idx ON appeals (submission_id);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    hackathon_id UUID REFERENCES hackathons(id) ON DELETE CASCADE,
    actor_id TEXT,
    action TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX audit_logs_hackathon_id_idx ON audit_logs (hackathon_id);
