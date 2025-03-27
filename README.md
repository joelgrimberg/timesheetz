# Timesheetz - write hours like a unicorn

<img src="docs/images/unicorn.jpg" height="150" />

## Description

Timesheetz is a timesheet management application with two interfaces:

- **Terminal User Interface (TUI)**: Add, view, and delete timesheet entries
  directly from your terminal.
- **REST API**: Programmatically manage timesheet entries through HTTP
  endpoints. Supports the same operations as the TUI.

The application stores all entries in a Sqlite database

<img src="docs/images/timesheet.png" width="750" />

## Installation

### Prerequisites

- Go 1.24.1 or higher
- For the email feature to work, you need to have to create an account at
  https://resend.com and create an API key. You can then use the API key on
  first launch of the application.

### Application Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/username/timesheetz.git
   ```

2. Initialize the database:
   ```bash
   ./go run timesheet --init
   ```

## Usage

Run the application in dev mode as a TUI with the following command:

```bash
make dev
```

The application uses keyboard shortcuts for navigation and actions. See the
[keyboard shortcuts guide](docs/shortcuts.md) for a comprehensive list of
available commands.

## Logging

Log files can be found in the `logs` directory: ~/.config/timesheet/logs

## TODO

- [ ] update Readme and setup github pages
- [ ] Build application in pipeline
- [ ] Add RayCast extension
