CREATE TABLE IF NOT EXISTS projects (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    folder_name VARCHAR(64) NOT NULL,
    name VARCHAR(64) NOT NULL,
    UNIQUE(folder_name));

CREATE TABLE IF NOT EXISTS units (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id INT NOT NULL,
    next_unit_id INT NULL,
    folder_name VARCHAR(64) NOT NULL,
    name VARCHAR(64) NOT NULL,
    CONSTRAINT fk_project FOREIGN KEY(project_id) REFERENCES projects(id),
    UNIQUE(project_id, folder_name));

CREATE TABLE IF NOT EXISTS tasks (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id INT NOT NULL,
    unit_id INT NOT NULL,
    folder_name VARCHAR(64) NOT NULL,
    name VARCHAR(64) NOT NULL,
    position INT NOT NULL,
    extention VARCHAR(10) NOT NULL,
    description VARCHAR(1024) NOT NULL,
    input VARCHAR(1024) NOT NULL,
    output VARCHAR(1024) NOT NULL,
    source_code TEXT NOT NULL,
    fixed_tests TEXT NOT NULL,
    score INT NOT NULL,
    UNIQUE(project_id, unit_id, folder_name),
    CONSTRAINT fk_project FOREIGN KEY(project_id) REFERENCES projects(id),
    CONSTRAINT fk_unit FOREIGN KEY(unit_id) REFERENCES units(id));

CREATE TABLE IF NOT EXISTS languages (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    extention VARCHAR(10) PRIMARY KEY,
    template TEXT NOT NULL);

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    email VARCHAR(128) NOT NULL,
    pass VARCHAR(256) NOT NULL,
    name VARCHAR(128) NOT NULL,
    is_suspended BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(email));

CREATE TABLE IF NOT EXISTS task_completions (
    user_id INT NOT NULL,
    task_id INT NOT NULL,
    completion FLOAT NOT NULL DEFAULT 0,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id),
    CONSTRAINT fk_task FOREIGN KEY(task_id) REFERENCES tasks(id),
    UNIQUE(user_id, task_id));

CREATE TABLE IF NOT EXISTS solutions(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    task_id INT NOT NULL,
    language_id INT NOT NULL,
    hash VARCHAR(128) NOT NULL,
    text TEXT NOT NULL,
    response VARCHAR(1024) NULL,
    received_times INT NOT NULL DEFAULT 0,
    CONSTRAINT fk_task FOREIGN KEY(task_id) REFERENCES tasks(id),
    CONSTRAINT fk_language FOREIGN KEY(language_id) REFERENCES langauges(id),
    last_used TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(task_id, hash));

CREATE TABLE IF NOT EXISTS solution_attempts(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INT NOT NULL,
    task_id INT NOT NULL,
    solution_id BIGINT NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id),
    CONSTRAINT fk_task FOREIGN KEY(task_id) REFERENCES tasks(id),
    CONSTRAINT fk_solution FOREIGN KEY(solution_id) REFERENCES solutions(id),
    dt TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

CREATE TABLE IF NOT EXISTS leaderboard (
    user_id INT NOT NULL,
    score FLOAT NOT NULL DEFAULT 0,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id));
