#!/bin/bash
#
# watch-hindsite - rebuild hindsite project when it changes
#
# Usage: watch-hindsite [PROJECT_DIR]
#

set -u  # No unbound variables.
set -e  # Exit on error.
PROJECT_DIR=.
if [[ $# > 0 ]]; then
    PROJECT_DIR=$1
else
    echo Usage: watch-hindsite.sh [PROJECT_DIR]
    exit 1
fi
WATCH_DIRS="$PROJECT_DIR/content $PROJECT_DIR/template"
echo Watching $WATCH_DIRS
echo Press Ctrl+C to stop
echo
hindsite build $PROJECT_DIR
echo
inotifywait -q -m -r -e modify,create,delete,move --timefmt %T --format "%T %e %f" $WATCH_DIRS | while read EVENT
do
    echo $EVENT
    case "$EVENT" in
    MODIFY*|CREATE*)
        hindsite build $PROJECT_DIR -incremental
        ;;
    *)
        hindsite build $PROJECT_DIR
        ;;
    esac
    echo
done