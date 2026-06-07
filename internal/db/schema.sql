CREATE TABLE IF NOT EXISTS Jobs (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    JobName TEXT NOT NULL, 
    Schedule TEXT NOT NULL,
    Host TEXT NOT NULL,
    JobType TEXT NOT NULL DEFAULT 'Bash',
    Commands TEXT NOT NULL,
    Env TEXT NOT NULL DEFAULT "",
    JobStatus TEXT NOT NULL DEFAULT 'Active',
    Created TEXT NOT NULL DEFAULT (datetime('now')),
    Updated TEXT NOT NULL DEFAULT (datetime('now')),
    LastRun TEXT NOT NULL DEFAULT 'Never',
    NextRun TEXT NOT NULL DEFAULT 'I should put some math here to calc out the next run or something'
);

CREATE TABLE IF NOT EXISTS Runs (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    JobID INTEGER FOREIGN KEY ON DELETE SET NULL,
    JobName TEXT NOT NULL,
    RunTimestamp TEXT NOT NULL,
    ExitCode INTEGER NOT NULL
)