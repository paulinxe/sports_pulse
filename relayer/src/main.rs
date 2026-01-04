use tokio_postgres::NoTls;
use std::error::Error;

mod test_blk_interaction;

enum ErrorCodes {
    DatabaseConnectionError = 1,
    MissingEnvironmentVariable = 2,
    QueryExecutionError = 3,
}

const SIGNED_MATCH_STATUS: i32 = 4;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let db_host = std::env::var("DB_HOST").unwrap_or_else(|_| {
        eprintln!("Error: DB_HOST environment variable is not set");
        std::process::exit(ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let db_port = std::env::var("DB_PORT").unwrap_or_else(|_| {
        eprintln!("Error: DB_PORT environment variable is not set");
        std::process::exit(ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let db_user = std::env::var("DB_USER").unwrap_or_else(|_| {
        eprintln!("Error: DB_USER environment variable is not set");
        std::process::exit(ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let db_password = std::env::var("DB_PASSWORD").unwrap_or_else(|_| {
        eprintln!("Error: DB_PASSWORD environment variable is not set");
        std::process::exit(ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let db_name = std::env::var("DB_NAME").unwrap_or_else(|_| {
        eprintln!("Error: DB_NAME environment variable is not set");
        std::process::exit(ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let connection_string = format!(
        "postgres://{}:{}@{}:{}/{}",
        db_user, db_password, db_host, db_port, db_name
    );
    
    let (client, connection) = tokio_postgres::connect(&connection_string, NoTls)
        .await
        .map_err(|e| {
            eprintln!("Error connecting to database: {}", e);
            std::process::exit(ErrorCodes::DatabaseConnectionError as i32);
        })?;

    // Spawn the connection in a new task so the main thread can continue
    // as this connection.await blocks the thread forever.
    tokio::spawn(async move {
        if let Err(e) = connection.await {
            // TODO: most probably here we should panic?
            eprintln!("Connection error: {}", e);
        }
    });

    let rows = client
        .query(
            "SELECT canonical_id, home_team_id, away_team_id, home_team_score, away_team_score, signature FROM matches WHERE status = $1",
            &[&SIGNED_MATCH_STATUS],
        )
        .await
        .map_err(|e| {
            eprintln!("Error executing query: {}", e);
            std::process::exit(ErrorCodes::QueryExecutionError as i32);
        })?;

    println!("Results:");
    for row in rows {
        let id: String = row.get("id");
        println!("id: {}", id);
    }

    test_blk_interaction::test_connection().await?;

    Ok(())
}
