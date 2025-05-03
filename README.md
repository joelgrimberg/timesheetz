# Timesheetz - write hours like a unicorn

<img src="docs/images/unicorn.jpg" height="150" />

## Description

Timesheetz is a timesheet management application with two interfaces:

- **Terminal User Interface (TUI)**: Add, view, and delete timesheet entries
  directly from your terminal.
- **REST API**: Programmatically manage timesheet entries through HTTP
  endpoints. Supports the same operations as the TUI.

The application stores all entries in a Sqlite database and features:

- 📅 Monthly calendar view with weekend indicators
- 📋 Copy/paste functionality with visual feedback
- 📊 Automatic total calculations
- 📤 Export to PDF or Excel
- 📧 Email integration via Resend.com
- 🔄 Real-time updates via API
- ⌨️ Vim-inspired keyboard shortcuts

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

Run the application as a background service with the --no-tui flag:

```bash
make dev ARGS="--no-tui"
```

The application uses keyboard shortcuts for navigation and actions. See the
[keyboard shortcuts guide](docs/shortcuts.md) for a comprehensive list of
available commands.

## Configuration

The application can be configured through `config.json`:

- Set document type (PDF/Excel) for exports
- Configure email settings (requires Resend.com API key)
- Enable/disable API server
- Set development mode to avoid cluttering production data

## Development

Within the config file, make sure to set mode to "development" to not clutter
your production data.

## Logging

Log files can be found in the `logs` directory: ~/.config/timesheet/logs

## API Documentation

The application provides a REST API for programmatic access. See the
[API documentation](docs/api.md) for available endpoints and examples.

## TODO

- [ ] update Readme and setup github pages
- [ ] Build application in pipeline
- [ ] Add RayCast extension
