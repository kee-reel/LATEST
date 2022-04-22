CREATE TABLE IF NOT EXISTS projects (
	id SERIAL PRIMARY KEY,
	folder_name VARCHAR(64) NOT NULL,
	name VARCHAR(64) NOT NULL,
	UNIQUE(folder_name));

CREATE TABLE IF NOT EXISTS units (
	id SERIAL PRIMARY KEY,
	project_id INTEGER NOT NULL,
	next_unit_id INTEGER NULL,
	folder_name VARCHAR(64) NOT NULL,
	name VARCHAR(64) NOT NULL,
	UNIQUE(folder_name, project_id));

CREATE TABLE IF NOT EXISTS tasks (
	id SERIAL PRIMARY KEY,
	project_id INTEGER NOT NULL,
	unit_id INTEGER NOT NULL,
	folder_name VARCHAR(64) NOT NULL,
	name VARCHAR(64) NOT NULL,
	position INTEGER NOT NULL,
	extention VARCHAR(10) NOT NULL,
	description VARCHAR(1024) NOT NULL,
	input VARCHAR(1024) NOT NULL,
	output VARCHAR(1024) NOT NULL,
	source_code TEXT NOT NULL,
	fixed_tests TEXT NOT NULL,
    score INT NOT NULL,
	UNIQUE(folder_name, project_id, unit_id));

CREATE TABLE IF NOT EXISTS solution_templates (
	extention VARCHAR(10) PRIMARY KEY,
	source_code TEXT NOT NULL);

CREATE TABLE IF NOT EXISTS solutions(
	user_id INTEGER NOT NULL,
	task_id INTEGER NOT NULL,
	completion FLOAT NOT NULL,
	dt TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

CREATE TABLE IF NOT EXISTS solutions_sources(
	user_id INTEGER NOT NULL,
	task_id INTEGER NOT NULL,
	source_code TEXT NOT NULL,
	PRIMARY KEY(user_id, task_id));

CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	email VARCHAR(128) NOT NULL,
	pass VARCHAR(256) NOT NULL,
	name VARCHAR(128) NOT NULL,
	UNIQUE(email));

CREATE TABLE IF NOT EXISTS leaderboard (
	user_id INTEGER NOT NULL,
	project_id INTEGER NOT NULL,
	score FLOAT NOT NULL DEFAULT 0,
	UNIQUE(user_id, project_id));
