CREATE TABLE IF NOT EXISTS Jobs (
    JobID INTEGER PRIMARY KEY AUTOINCREMENT,
    JobName TEXT NOT NULL, 
    Schedule TEXT NOT NULL,
    Host TEXT NOT NULL,
    JobStatus TEXT NOT NULL DEFAULT 'Active',
    JobType TEXT NOT NULL DEFAULT 'Bash',
    Commands TEXT NOT NULL,
    Created TEXT NOT NULL DEFAULT (datetime('now')),
    Updated TEXT NOT NULL DEFAULT (datetime('now'))
);
