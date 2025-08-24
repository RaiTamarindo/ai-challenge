#!/bin/sh
# Database and user setup script for feature voting platform
# This script creates the application user and grants necessary permissions

set -e

# Use environment variables with defaults
POSTGRES_ADMIN_USERNAME=${POSTGRES_ADMIN_USERNAME:-postgres}
POSTGRES_ADMIN_PASSWORD=${POSTGRES_ADMIN_PASSWORD:-postgres_admin_pass}
POSTGRES_STANDARD_USERNAME=${POSTGRES_STANDARD_USERNAME:-voting_app}
POSTGRES_STANDARD_PASSWORD=${POSTGRES_STANDARD_PASSWORD:-voting_app_pass}
POSTGRES_DB=${POSTGRES_DB:-feature_voting_platform}

echo "Setting up database users and permissions..."

# Function to execute SQL commands
execute_sql() {
    PGPASSWORD=$POSTGRES_ADMIN_PASSWORD psql -h localhost -U $POSTGRES_ADMIN_USERNAME -d $POSTGRES_DB -c "$1"
}

# Check if standard user exists, create if not
echo "Checking if user '$POSTGRES_STANDARD_USERNAME' exists..."
USER_EXISTS=$(PGPASSWORD=$POSTGRES_ADMIN_PASSWORD psql -h localhost -U $POSTGRES_ADMIN_USERNAME -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$POSTGRES_STANDARD_USERNAME'")

if [ "$USER_EXISTS" != "1" ]; then
    echo "Creating user '$POSTGRES_STANDARD_USERNAME'..."
    PGPASSWORD=$POSTGRES_ADMIN_PASSWORD psql -h localhost -U $POSTGRES_ADMIN_USERNAME -d postgres -c "
        CREATE USER $POSTGRES_STANDARD_USERNAME WITH PASSWORD '$POSTGRES_STANDARD_PASSWORD';
    "
    echo "User '$POSTGRES_STANDARD_USERNAME' created successfully."
else
    echo "User '$POSTGRES_STANDARD_USERNAME' already exists."
fi

# Grant permissions to the standard user on the database
echo "Granting permissions to user '$POSTGRES_STANDARD_USERNAME'..."
execute_sql "GRANT CONNECT ON DATABASE $POSTGRES_DB TO $POSTGRES_STANDARD_USERNAME;"
execute_sql "GRANT USAGE ON SCHEMA public TO $POSTGRES_STANDARD_USERNAME;"
execute_sql "GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO $POSTGRES_STANDARD_USERNAME;"
execute_sql "GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO $POSTGRES_STANDARD_USERNAME;"

# Grant permissions for future tables (important for migrations)
execute_sql "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO $POSTGRES_STANDARD_USERNAME;"
execute_sql "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO $POSTGRES_STANDARD_USERNAME;"

echo "Database setup completed successfully!"
echo "Application user '$POSTGRES_STANDARD_USERNAME' is ready to use."

# Display connection information
echo ""
echo "Connection information:"
echo "  Database: $POSTGRES_DB"
echo "  Admin user: $POSTGRES_ADMIN_USERNAME"
echo "  App user: $POSTGRES_STANDARD_USERNAME"
echo "  Host: localhost"
echo "  Port: 5432"
echo ""
echo "Connection string for Go application:"
echo "  postgresql://$POSTGRES_STANDARD_USERNAME:$POSTGRES_STANDARD_PASSWORD@localhost:5432/$POSTGRES_DB?sslmode=disable"