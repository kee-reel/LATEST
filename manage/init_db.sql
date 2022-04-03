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
    score INT NOT NULL,
	UNIQUE(folder_name, project_id, unit_id));

CREATE TABLE IF NOT EXISTS solution_templates (
	extention VARCHAR(10) PRIMARY KEY,
	source_code TEXT NOT NULL);

CREATE TABLE IF NOT EXISTS solutions(
	user_id INTEGER NOT NULL,
	task_id INTEGER NOT NULL,
	is_passed BOOLEAN NOT NULL,
	dt TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

CREATE TABLE IF NOT EXISTS solutions(
    project_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	score INTEGER NOT NULL DEFAULT 0,
    Primary KEY(project_id, user_id));

CREATE TABLE IF NOT EXISTS solutions_sources(
	user_id INTEGER NOT NULL,
	task_id INTEGER NOT NULL,
	source_code TEXT NOT NULL,
	PRIMARY KEY(token_id, task_id));

CREATE TABLE IF NOT EXISTS tokens (
	id SERIAL PRIMARY KEY,
	token VARCHAR(256) NOT NULL,
	user_id INTEGER NOT NULL,
	ip VARCHAR(15) NOT NULL,
	UNIQUE(token),
	UNIQUE(user_id, ip));

CREATE TABLE IF NOT EXISTS registration_tokens (
	token VARCHAR(256) PRIMARY KEY,
	email VARCHAR(128) NOT NULL,
	ip VARCHAR(15) NOT NULL,
	pass VARCHAR(256) NOT NULL,
	name VARCHAR(50) NOT NULL,
    UNIQUE(email, ip));

CREATE TABLE IF NOT EXISTS verification_tokens (
	token VARCHAR(256) PRIMARY KEY,
	ip VARCHAR(15) NOT NULL,
	user_id INTEGER NOT NULL,
    UNIQUE(ip, user_id));

CREATE TABLE IF NOT EXISTS restore_tokens (
	token VARCHAR(256) PRIMARY KEY,
	ip VARCHAR(15) NOT NULL,
	user_id INTEGER NOT NULL,
	pass VARCHAR(256) NOT NULL,
    UNIQUE(user_id, ip));

CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	email VARCHAR(128) NOT NULL,
	pass VARCHAR(256) NOT NULL,
	name VARCHAR(128) NOT NULL,
	UNIQUE(email));

