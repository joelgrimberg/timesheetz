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
- 🔄 Automatic startup on system boot
- 🛠️ Easy management with control scripts

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

2. Build the application for your platform:

   ```bash
   ./scripts/build.sh
   ```

3. Install the application:

   For macOS:
   ```bash
   ./scripts/install-mac.sh
   ```

   For Windows (run in PowerShell as administrator):
   ```powershell
   .\scripts\install-win.ps1
   ```

   For Linux:
   ```bash
   ./scripts/install-linux.sh
   ```

The installation script will:
- Install the application in the appropriate system directory
- Set up automatic startup on system boot
- Configure necessary permissions
- Start the application

## Managing the Application

The application comes with a management script that provides easy control over the application lifecycle:

```bash
# Check application status
./scripts/manage.sh status

# Stop the application
./scripts/manage.sh stop

# Start the application
./scripts/manage.sh start

# Restart the application
./scripts/manage.sh restart

# Reload the application (graceful reload)
./scripts/manage.sh reload
```

### Updating the Application

To update to a newer version:

1. Stop the current version:
   ```bash
   ./scripts/manage.sh stop
   ```

2. Build and install the new version:
   ```bash
   ./scripts/build.sh
   ./scripts/install-mac.sh  # or install-win.ps1 or install-linux.sh
   ```

3. Start the new version:
   ```bash
   ./scripts/manage.sh start
   ```

Or simply use the restart command:
```bash
./scripts/manage.sh restart
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

### Command-line Flags

The application supports the following command-line flags:

- `--no-tui`: Run only the API server without the TUI
- `--port <number>`: Specify the port for the API server (default: 8080)
- `--dev`: Run in development mode (uses local database)
- `--init`: Initialize the database
- `--help`: Show help message
- `--verbose`: Show detailed output

Example:
```bash
# Run API server on port 3000 in development mode
./timesheet --no-tui --port 3000 --dev

# Show help message
./timesheet --help
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

Log files can be found in the following locations:
- macOS: `$HOME/Applications/Timesheetz/error.log` and `output.log`
- Windows: `%LOCALAPPDATA%\Timesheetz\`
- Linux: Use `journalctl --user -u timesheetz.service`

## API Documentation

The application provides a REST API for programmatic access. See the
[API documentation](docs/api.md) for available endpoints and examples.

## TODO

- [ ] update Readme and setup github pages
- [ ] Build application in pipeline
- [ ] Add RayCast extension
