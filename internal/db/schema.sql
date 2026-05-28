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
    LastRun TEXT DEFAULT 'Never',
    NextRun TEXT DEFAULT 'I should put some math here to calc out the next run or something'
);
