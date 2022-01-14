#!/bin/bash

dbname='./tasks.db'

sqlite3 $dbname 'CREATE TABLE IF NOT EXISTS subject ( id INTEGER PRIMARY KEY, name VARCHAR(64))'
sqlite3 $dbname 'CREATE TABLE IF NOT EXISTS work (
	id INTEGER,
	subject INTEGER,
	next_work_id INTEGER NULL,
	name VARCHAR(64),
	PRIMARY KEY(id, subject))'
sqlite3 $dbname 'CREATE TABLE IF NOT EXISTS variant (
	id INTEGER,
	subject INTEGER,
	work INTEGER,
	name VARCHAR(64),
	PRIMARY KEY(id, subject, work))'
sqlite3 $dbname 'CREATE TABLE IF NOT EXISTS task (
	id INTEGER PRIMARY KEY,
	subject INTEGER,
	work INTEGER,
	variant INTEGER,
	number INTEGER,
	name VARCHAR(64),
	extention VARCHAR(10),
	desc VARCHAR(1024),
	input VARCHAR(512),
	output VARCHAR(128),
	UNIQUE(subject, work, variant, number))'
sqlite3 $dbname 'CREATE TABLE IF NOT EXISTS solution(
	token_id INTEGER,
	task_id INTEGER,
	is_user_tests_passed BOOLEAN,
	is_passed BOOLEAN,
	dt TIMESTAMP DEFAULT CURRENT_TIMESTAMP)'
sqlite3 $dbname 'CREATE TABLE IF NOT EXISTS access_token (
	id INTEGER PRIMARY KEY,
	token VARCHAR(256),
	user_id INTEGER,
	subject INTEGER,
	variant INTEGER,
	UNIQUE(token))'
sqlite3 $dbname 'CREATE TABLE IF NOT EXISTS user (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	group_name VARCHAR(128),
	number INTEGER,
	name VARCHAR(128),
	last_name VARCHAR(128),
	UNIQUE(group_name, number))'
