#!/bin/bash

# Function to find the process ID of the running application
find_pid() {
    # Look for the process in the Applications directory
    if [ -f "$HOME/Applications/Timesheetz/timesheet" ]; then
        pgrep -f "$HOME/Applications/Timesheetz/timesheet"
    else
        # Fallback to looking in the current directory
        pgrep -f "timesheet"
    fi
}

# Function to stop the application
stop_app() {
    local pid=$(find_pid)
    if [ -n "$pid" ]; then
        echo "Stopping Timesheetz (PID: $pid)..."
        kill $pid
        # Wait for the process to stop
        while kill -0 $pid 2>/dev/null; do
            sleep 1
        done
        echo "Timesheetz stopped."
    else
        echo "Timesheetz is not running."
    fi
}

# Function to start the application
start_app() {
    if [ -f "$HOME/Applications/Timesheetz/timesheet" ]; then
        echo "Starting Timesheetz from Applications directory..."
        "$HOME/Applications/Timesheetz/timesheet" &
    else
        echo "Starting Timesheetz from current directory..."
        # Get the directory where this script is located
        SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
        
        # Change to project root directory
        cd "$PROJECT_ROOT" || {
            echo "Error: Could not change to project directory: $PROJECT_ROOT"
            exit 1
        }
        
        # Check if the binary exists
        if [ ! -f "./bin/timesheet" ]; then
            echo "Error: Binary not found at ./bin/timesheet"
            echo "Please run ./scripts/build.sh first to build the application"
            exit 1
        fi
        
        # Start the application
        ./bin/timesheet &
    fi
    echo "Timesheetz started."
}

# Function to reload the application
reload_app() {
    local pid=$(find_pid)
    if [ -n "$pid" ]; then
        echo "Reloading Timesheetz (PID: $pid)..."
        # Send SIGHUP signal for graceful reload
        kill -HUP $pid
        echo "Timesheetz reloaded."
    else
        echo "Timesheetz is not running. Starting it..."
        start_app
    fi
}

# Function to check application status
status_app() {
    local pid=$(find_pid)
    if [ -n "$pid" ]; then
        echo "Timesheetz is running (PID: $pid)"
        # Check if the API is responding
        if curl -s http://localhost:8080/health > /dev/null; then
            echo "API is responding"
        else
            echo "API is not responding"
        fi
    else
        echo "Timesheetz is not running"
    fi
}

# Main script logic
case "$1" in
    start)
        start_app
        ;;
    stop)
        stop_app
        ;;
    restart)
        stop_app
        sleep 2
        start_app
        ;;
    reload)
        reload_app
        ;;
    status)
        status_app
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|reload|status}"
        exit 1
        ;;
esac

exit 0 