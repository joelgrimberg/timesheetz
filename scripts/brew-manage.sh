#!/bin/bash

# Timesheetz Launch Agent Management Script
# Manages the macOS Launch Agent for the Homebrew-installed Timesheetz

PLIST_PATH=~/Library/LaunchAgents/com.timesheetz.plist

case "$1" in
    start)
        echo "Starting Timesheetz Launch Agent..."
        launchctl load "$PLIST_PATH"
        sleep 1
        if launchctl list | grep -q com.timesheetz; then
            echo "✓ Started successfully"
        else
            echo "✗ Failed to start"
            exit 1
        fi
        ;;

    stop)
        echo "Stopping Timesheetz Launch Agent..."
        launchctl unload "$PLIST_PATH"
        echo "✓ Stopped"
        ;;

    restart)
        echo "Restarting Timesheetz Launch Agent..."
        launchctl unload "$PLIST_PATH" 2>/dev/null || true
        launchctl load "$PLIST_PATH"
        sleep 1
        if launchctl list | grep -q com.timesheetz; then
            echo "✓ Restarted successfully"
        else
            echo "✗ Failed to restart"
            exit 1
        fi
        ;;

    status)
        if launchctl list | grep -q com.timesheetz; then
            PID=$(launchctl list | grep com.timesheetz | awk '{print $1}')
            echo "✓ Running (PID: $PID)"
        else
            echo "✗ Not running"
        fi
        ;;

    logs)
        echo "Tailing output logs (Ctrl+C to exit)..."
        tail -f ~/Library/Logs/timesheetz.out
        ;;

    errors)
        echo "Tailing error logs (Ctrl+C to exit)..."
        tail -f ~/Library/Logs/timesheetz.err
        ;;

    *)
        echo "Timesheetz Launch Agent Management"
        echo ""
        echo "Usage: $0 {start|stop|restart|status|logs|errors}"
        echo ""
        echo "Commands:"
        echo "  start    - Load and start the Launch Agent"
        echo "  stop     - Unload and stop the Launch Agent"
        echo "  restart  - Restart the Launch Agent"
        echo "  status   - Check if the Launch Agent is running"
        echo "  logs     - Tail the output logs"
        echo "  errors   - Tail the error logs"
        exit 1
        ;;
esac
