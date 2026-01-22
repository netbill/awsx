CREATE TABLE upload_sessions (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,

    purpose TEXT NOT NULL,                 -- "review_photos", "place_gallery", ...
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    expires_at TIMESTAMPTZ NOT NULL,

    max_files INT NOT NULL CHECK (max_files > 0)
);

CREATE INDEX upload_sessions_owner_idx  ON upload_sessions(owner_id);
CREATE INDEX upload_sessions_expires_idx ON upload_sessions(expires_at);

CREATE TABLE upload_files (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES upload_sessions(id) ON DELETE CASCADE,

    object_key TEXT NOT NULL UNIQUE,       -- staging key
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'UTC')
);

CREATE INDEX upload_files_session_idx ON upload_files(session_id);
