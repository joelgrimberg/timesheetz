# Timesheet Application Keyboard Shortcuts

Below is a comprehensive list of keyboard shortcuts available in the timesheet
application:

| Shortcut   | Description                    |
| ---------- | ------------------------------ |
| ↑ / k      | Move cursor up                 |
| ↓ / j      | Move down                      |
| ← / h      | Go to previous month           |
| → / l      | Go to next month               |
| t          | Jump to today's date           |
| Enter      | Select/edit entry              |
| a          | Add a new entry                |
| c          | Clear the selected entry       |
| y          | Yank (copy) the selected entry |
| p          | Paste previously yanked entry  |
| u          | Jump up multiple rows          |
| d          | Jump down multiple rows        |
| P          | Print timesheet to PDF         |
| S          | Send timesheet via email       |
| ?          | Toggle help view               |
| q / Ctrl+C | Quit application               |
| Esc        | Clear yanked entry             |

## Navigation Tips

- Use **h** and **l** to navigate between months
- Use **t** to quickly return to today's date regardless of which month you're
  viewing
- The **u** and **d** keys allow for faster navigation through long timesheets
- Weekend days are visually marked with a 💤 emoji for easy identification

## Copy & Paste Workflow

1. Navigate to an entry you want to copy
2. Press **y** to yank (copy) the entry - the row will be highlighted in green
3. Move to the date where you want to duplicate the entry
4. Press **p** to paste the entry
5. Press **Esc** to clear the yanked entry and remove the green highlight

## Form Mode Navigation

When adding or editing an entry:
- Use **Tab** to move to the next field
- Use **Shift+Tab** to move to the previous field
- Press **Enter** on the last field to submit
- Press **Esc** to cancel and return to the timesheet view

## Input Requirements

- Date format must be YYYY-MM-DD (e.g., 2024-03-20)
- Hours must be non-negative whole numbers
- Client name is required
- Total hours are automatically calculated

## Export Options

- **P** - Generate a PDF of the current timesheet view
- **S** - Generate a PDF and send it via email
- Document type (PDF/Excel) can be configured in `config.json`

## API Integration

The application supports real-time updates when entries are modified via the API:
- Changes are immediately reflected in the UI
- The current view and cursor position are preserved during updates
- See the [API documentation](api.md) for available endpoints

## Configuration

The application can be configured through `config.json`:
- Set document type (PDF/Excel) for exports
- Configure email settings (requires Resend.com API key)
- Enable/disable API server
- Set development mode to avoid cluttering production data
