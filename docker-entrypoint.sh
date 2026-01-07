#!/bin/sh
set -e

# Default PUID/PGID to appuser's values if not set
PUID=${PUID:-10001}
PGID=${PGID:-10001}

echo "
───────────────────────────────────────
User UID:    $(id -u appuser)
User GID:    $(id -g appuser)
───────────────────────────────────────
"

# Check if we need to change UID/GID
if [ "$(id -u appuser)" != "$PUID" ] || [ "$(id -g appuser)" != "$PGID" ]; then
    echo "Adjusting appuser UID/GID to $PUID:$PGID"

    # Change GID
    if [ "$(id -g appuser)" != "$PGID" ]; then
        groupmod -o -g "$PGID" appuser
    fi

    # Change UID
    if [ "$(id -u appuser)" != "$PUID" ]; then
        usermod -o -u "$PUID" appuser
    fi

    echo "
───────────────────────────────────────
Updated UID: $(id -u appuser)
Updated GID: $(id -g appuser)
───────────────────────────────────────
"
fi

# Fix ownership of config directory if it exists
if [ -d "/home/appuser/.config/anilist-mal-sync" ]; then
    echo "Fixing ownership of /home/appuser/.config/anilist-mal-sync"
    chown -R appuser:appuser /home/appuser/.config/anilist-mal-sync
fi

# Execute command as appuser
echo "Starting application as appuser..."
exec su-exec appuser "$@"
