CREATE TABLE IF NOT EXISTS fax_jobs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject         TEXT NOT NULL DEFAULT '',
    destination_ext INTEGER NOT NULL,
    original_file   TEXT NOT NULL,
    tiff_file       TEXT NOT NULL DEFAULT '',
    call_file       TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'queued' CHECK(status IN ('converting','queued','attempted','failed')),
    error_message   TEXT NOT NULL DEFAULT '',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fax_jobs_user_id ON fax_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_fax_jobs_created_at ON fax_jobs(created_at);
