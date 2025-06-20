# Vacation Tracking

The vacation tracking feature allows you to monitor your vacation hours throughout the year. It helps you keep track of:
- Total vacation hours taken
- Remaining vacation hours
- Vacation history
- Yearly vacation targets

## Configuration

The vacation tracking system is configured with:
- A yearly target of vacation hours (default: 180 hours)
- A category for vacation entries (default: "Vacation")

You can modify these settings during the initial setup or through the Settings tab.

## Using the Vacation Tab

The Vacation tab provides a user-friendly interface to:
1. View your vacation history
2. Add new vacation entries
3. Track your progress against the yearly target
4. Switch between different years

### Adding a Vacation Entry

To add a new vacation entry:
1. Select the "Vacation" tab
2. Fill in the form at the bottom of the screen:
   - Date: The date of your vacation
   - Hours: Number of hours taken
   - Notes: Optional notes about the vacation
3. Click "Add" to save the entry

### Viewing Vacation History

The main table shows:
- All vacation entries for the selected year
- Total hours taken
- Remaining hours
- Yearly target

You can switch between years using the dropdown at the top of the screen.

## API Endpoints

The vacation tracking system provides the following API endpoints:

- `GET /api/vacation`: Get all vacation entries for a year
- `POST /api/vacation`: Create a new vacation entry
- `PUT /api/vacation/:id`: Update an existing vacation entry
- `DELETE /api/vacation/:id`: Delete a vacation entry

## Database Schema

The vacation data is stored in the `vacation` table with the following structure:

```sql
CREATE TABLE vacation (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    hours INTEGER NOT NULL,
    category TEXT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
``` 