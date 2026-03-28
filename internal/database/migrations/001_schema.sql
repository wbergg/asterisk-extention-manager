CREATE TABLE IF NOT EXISTS users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT NOT NULL UNIQUE,
    password        TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin','user')),
    min_ext         INTEGER NOT NULL DEFAULT 0,
    max_ext         INTEGER NOT NULL DEFAULT 0,
    call_log_access BOOLEAN NOT NULL DEFAULT 1,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS extensions (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    extension      INTEGER NOT NULL UNIQUE,
    user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sip_username   TEXT NOT NULL UNIQUE,
    sip_password   TEXT NOT NULL,
    callerid       TEXT NOT NULL DEFAULT '',
    context        TEXT NOT NULL DEFAULT 'internal',
    directory_only INTEGER NOT NULL DEFAULT 0,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    action     TEXT NOT NULL,
    detail     TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS blocked_extensions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    extension  INTEGER NOT NULL UNIQUE,
    reason     TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS fax_jobs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject         TEXT NOT NULL DEFAULT '',
    destination_ext INTEGER NOT NULL,
    original_file   TEXT NOT NULL,
    tiff_file       TEXT NOT NULL DEFAULT '',
    call_file       TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'queued' CHECK(status IN ('converting','queued','attempted','sent','failed')),
    error_message   TEXT NOT NULL DEFAULT '',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_extensions_user_id ON extensions(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_fax_jobs_user_id ON fax_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_fax_jobs_created_at ON fax_jobs(created_at);
