use tokio_postgres::NoTls;
use std::error::Error;
use log::{info, error};

pub async fn init() -> Result<tokio_postgres::Client, Box<dyn Error>> {
    let db_host = std::env::var("DB_HOST").unwrap_or_else(|_| {
        eprintln!("Error: DB_HOST environment variable is not set");
        std::process::exit(crate::ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let db_port = std::env::var("DB_PORT").unwrap_or_else(|_| {
        eprintln!("Error: DB_PORT environment variable is not set");
        std::process::exit(crate::ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let db_user = std::env::var("DB_USER").unwrap_or_else(|_| {
        eprintln!("Error: DB_USER environment variable is not set");
        std::process::exit(crate::ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let db_password = std::env::var("DB_PASSWORD").unwrap_or_else(|_| {
        eprintln!("Error: DB_PASSWORD environment variable is not set");
        std::process::exit(crate::ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let db_name = std::env::var("DB_NAME").unwrap_or_else(|_| {
        eprintln!("Error: DB_NAME environment variable is not set");
        std::process::exit(crate::ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let connection_string = format!(
        "postgres://{}:{}@{}:{}/{}",
        db_user, db_password, db_host, db_port, db_name
    );
    
    info!("Connecting to PostgreSQL database:");
    
    let (client, connection) = tokio_postgres::connect(&connection_string, NoTls)
        .await
        .map_err(|e| {
            error!("Failed to connect to PostgreSQL database:");
            error!("  Host: {}", db_host);
            error!("  Port: {}", db_port);
            error!("  User: {}", db_user);
            error!("  Database: {}", db_name);
            error!("  Connection string (without password): postgres://{}:***@{}:{}/{}", db_user, db_host, db_port, db_name);
            error!("  Error type: {:?}", e);
            error!("  Error message: {}", e);
            error!("  Error source: {:?}", e.source());
            std::process::exit(crate::ErrorCodes::DatabaseConnectionError as i32);
        })?;
    
    info!("Database connection established successfully");

    // Spawn the connection in a new task so the main thread can continue
    // as this connection.await blocks the thread forever.
    tokio::spawn(async move {
        if let Err(e) = connection.await {
            error!("Database connection error occurred:");
            error!("  Error type: {:?}", e);
            error!("  Error message: {}", e);
            error!("  Error source: {:?}", e.source());
            error!("  Note: The connection task has terminated. The application may not be able to execute queries.");
            // TODO: most probably here we should panic?
        }
    });

    Ok(client)
}
