#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
# set -o xtrace

# This script acts as a robust PID 1 wrapper, handling SIGTERM/SIGINT for graceful
# shutdown and SIGCHLD to prevent zombie processes. It is suitable for wrapping
# applications that may fork their own child processes.

# Global variables
child_pid=""
shutdown_requested=false

# A function to reap zombie processes.
# It loops and calls `wait -n` which waits for the next child to terminate.
# This loop is necessary to handle multiple children exiting in quick succession,
# as the SIGCHLD signal may not be delivered for each one.
reap_children() {
    # Don't reap during shutdown to avoid interfering with main process cleanup
    if [ "$shutdown_requested" = "true" ]; then
        return
    fi
    
    # The `while` loop continues as long as `wait -n` successfully reaps a child.
    # `wait -n` returns the exit code of the reaped child or >128 if there are no children.
    # We suppress errors because `wait -n` may fail if no children are left.
    while wait -n -p pid &> /dev/null; do
        exit_code=$?
        # Skip reaping our main child process - let the main wait handle it
        if [ "$pid" -eq "$child_pid" ] 2>/dev/null; then
            continue
        fi
        if [ "$pid" -ne 0 ]; then
            echo "Reaped orphaned child PID=$pid with exit code: $exit_code"
        fi
    done
}

# A function to handle the shutdown of the main application.
shutdown() {
    echo "Signal received, shutting down main application..."
    shutdown_requested=true
    
    if [ -n "$child_pid" ] && kill -0 "$child_pid" 2>/dev/null; then
        # Try to send SIGTERM to the process group first (negative PID)
        # This works if the child process is a process group leader
        if kill -TERM -- "-$child_pid" 2>/dev/null; then
            echo "Sent SIGTERM to process group of child PID=$child_pid."
        else
            # Fallback: send to the child process directly
            kill -TERM -- "$child_pid" 2>/dev/null || true
            echo "Sent SIGTERM to child PID=$child_pid."
        fi
        
        # Also try to kill any children of the main process
        # Get all child processes and send them SIGTERM
        if command -v pgrep >/dev/null 2>&1; then
            child_pids=$(pgrep -P "$child_pid" 2>/dev/null || true)
            if [ -n "$child_pids" ]; then
                echo "Sending SIGTERM to child processes: $child_pids"
                echo "$child_pids" | xargs -r kill -TERM 2>/dev/null || true
            fi
        fi
        
        # Give the process time to shut down gracefully
        local timeout=10
        while [ $timeout -gt 0 ] && kill -0 "$child_pid" 2>/dev/null; do
            sleep 1
            timeout=$((timeout - 1))
        done
        
        # Force kill if still running
        if kill -0 "$child_pid" 2>/dev/null; then
            echo "Child process did not exit gracefully, sending SIGKILL..."
            kill -KILL -- "$child_pid" 2>/dev/null || true
        fi
    fi
}

# Trap signals:
# - SIGCHLD: A child process has terminated. Run our reaper function.
# - SIGTERM, SIGINT: Request to shut down the container. Run our shutdown function.
trap reap_children CHLD
trap shutdown TERM INT

# Start the main application in the background
# We'll handle process group signaling in the shutdown function
"$@" &
child_pid=$!
echo "Main application started with child PID=$child_pid"

# Main event loop - wait for the main application process to exit
# We use a loop with short waits to ensure signal responsiveness
while kill -0 "$child_pid" 2>/dev/null; do
    if [ "$shutdown_requested" = "true" ]; then
        break
    fi
    # Short wait to allow signal processing
    sleep 0.1
done

# Get the exit code of the main application
main_app_exit_code=0
if kill -0 "$child_pid" 2>/dev/null; then
    # Process is still running, wait for it to finish
    wait "$child_pid" 2>/dev/null
    main_app_exit_code=$?
else
    # Process already exited, try to get its exit code
    wait "$child_pid" 2>/dev/null
    main_app_exit_code=$?
fi

echo "Main application (PID=$child_pid) exited with code: $main_app_exit_code."

# Clean up any remaining children
echo "Cleaning up remaining child processes..."
reap_children

# Exit the wrapper script with the same code as the main application.
exit "$main_app_exit_code"
