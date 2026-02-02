#!/bin/sh

# Auto-create superuser if env vars are present
if [ -n "$PB_ADMIN_EMAIL" ] && [ -n "$PB_ADMIN_PASSWORD" ]; then
    echo "Attempting to create superuser: $PB_ADMIN_EMAIL"
    /pb superuser upsert "$PB_ADMIN_EMAIL" "$PB_ADMIN_PASSWORD" || true
fi

# Start PocketBase
exec /pb serve --http=0.0.0.0:8090
