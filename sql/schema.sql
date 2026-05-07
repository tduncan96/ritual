CREATE TABLE IF NOT EXISTS jobs (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    JobName TEXT NOT NULL, 
    Schedule TEXT NOT NULL,
    Host TEXT NOT NULL,
    JobStatus TEXT NOT NULL DEFAULT 'Active',
    JobType TEXT NOT NULL DEFAULT 'Bash',
    Commands TEXT NOT NULL,
    Created TEXT NOT NULL DEFAULT (datetime('now')),
    Updated TEXT NOT NULL DEFAULT (datetime('now')),
    LastRun TEXT DEFAULT NULL,
    NextRun TEXT DEFAULT NULL
);
