-- SQLite doesn't support ALTER CHECK, so we recreate the table
CREATE TABLE IF NOT EXISTS fax_jobs_new (
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

INSERT OR IGNORE INTO fax_jobs_new SELECT * FROM fax_jobs;
DROP TABLE IF EXISTS fax_jobs;
ALTER TABLE fax_jobs_new RENAME TO fax_jobs;

CREATE INDEX IF NOT EXISTS idx_fax_jobs_user_id ON fax_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_fax_jobs_created_at ON fax_jobs(created_at);
