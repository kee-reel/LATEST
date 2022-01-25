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
	input VARCHAR(512) NOT NULL,
	output VARCHAR(128) NOT NULL,
	source_code TEXT NOT NULL,
	fixed_tests TEXT NOT NULL,
	UNIQUE(folder_name, project_id, unit_id));

CREATE TABLE IF NOT EXISTS solutions(
	token_id INTEGER NOT NULL,
	task_id INTEGER NOT NULL,
	is_passed BOOLEAN NOT NULL,
	dt TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

CREATE TABLE IF NOT EXISTS solutions_sources(
	token_id INTEGER NOT NULL,
	task_id INTEGER NOT NULL,
	source_code TEXT NOT NULL,
	PRIMARY KEY(token_id, task_id));

CREATE TABLE IF NOT EXISTS tokens (
	id SERIAL PRIMARY KEY,
	token VARCHAR(256) NOT NULL,
	user_id INTEGER NOT NULL,
	project_id INTEGER NOT NULL,
	UNIQUE(token));

CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	email VARCHAR(128) NOT NULL,
	nick VARCHAR(128) NOT NULL,
	UNIQUE(email, nick));

