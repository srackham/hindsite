#!/bin/bash
#
# watch-hindsite - rebuild hindsite project when it changes
#
# Usage: watch-hindsite PROJECT_DIR [OPTIONS]
#
# Performs incremental build if files are created or modified,
# Performs full rebuild if files are deleted or renamed,
#

set -u  # No unbound variables.
set -e  # Exit on error.
if [[ $# == 0 ]]; then
    echo Usage: watch-hindsite.sh PROJECT_DIR [OPTIONS]
    exit 1
fi
WATCH_DIRS="$1/content $1/template"
echo Watching $WATCH_DIRS
echo Press Ctrl+C to stop
echo
hindsite build "$@"
echo
while true; do
    EVENT=$(inotifywait -q -r -e modify,create,delete,move --format "%e: %f" $WATCH_DIRS)
    sleep 0.2s  # Allow some time for all editor saves to complete.
    echo $EVENT
    set +e      # Do not exit if there are build errors.
    case "$EVENT" in
    MODIFY*|CREATE*)
        hindsite build "$@" -incremental
        ;;
    *)
        hindsite build "$@"
        ;;
    esac
    set -e
    echo
done