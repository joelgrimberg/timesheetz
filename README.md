# Timesheetz - A unicorny way to track your hours

<img src="docs/images/unicorn.jpg" height="150" />

## Description

A simple timesheet app , written in GO, using a MySQL database.

Documentation regarding the api can be found [here](./docs/api.md)

## Installation

'timesheet --init' will initialize the database

## Usage

Run the application in dev mode as a TUI with the following command:

```bash

export DBUSER="root"
export DBPASSWORD="password"

make dev
```

## TODO

- [x] Build Table setup
- [x] add data from database to table
- [x] switch months in table view
- [x] add keys to the table
  - [x] add q
  - [x] add enter
  - [x] add up / down
  - [x] add extended key view
  - [x] add left / right keys to move between months
- [x] Create insert form
- [x] Create edit form
- [x] Add all database fields to insert form
- [x] Add all database fields to table
- [x] fix api server
- [ ] update Readme and setup github pages
- [ ] Build application in pipeline
- [ ] Create installer for database
- [ ] Add RayCast extension
