-- Create keycloak database
CREATE DATABASE keycloak;

-- Grant privileges to postgres user
GRANT ALL PRIVILEGES ON DATABASE keycloak TO postgres;
