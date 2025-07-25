#!/bin/bash

# Database Migration Script for Vistara AI
# This script applies all migrations in order

set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-vistara_ai}"
DB_USER="${DB_USER:-vistara}"
DB_PASSWORD="${DB_PASSWORD:-vistara123}"

MIGRATION_DIR="$(dirname "$0")/migrations"

echo "🚀 Starting database migration..."
echo "📍 Database: $DB_HOST:$DB_PORT/$DB_NAME"

# Function to run a migration file
run_migration() {
    local file=$1
    local filename=$(basename "$file")
    
    echo "🔄 Applying migration: $filename"
    
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -f "$file"
    
    if [ $? -eq 0 ]; then
        echo "✅ Migration $filename completed successfully"
    else
        echo "❌ Migration $filename failed"
        exit 1
    fi
}

# Check if database exists and is accessible
echo "🔌 Testing database connection..."
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "SELECT 'Database connection successful!' as status;" > /dev/null

if [ $? -ne 0 ]; then
    echo "❌ Cannot connect to database. Please check your configuration."
    exit 1
fi

echo "✅ Database connection successful"

# Apply migrations in order
for migration_file in "$MIGRATION_DIR"/*.sql; do
    if [ -f "$migration_file" ]; then
        run_migration "$migration_file"
    fi
done

echo ""
echo "🎉 All migrations completed successfully!"
echo ""
echo "📊 Database Summary:"
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "\dt"

echo ""
echo "🗺️ PostGIS Version:"
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "SELECT PostGIS_Version();"
