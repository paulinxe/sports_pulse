use std::error::Error;
use std::io::Error as IOError;
use std::io::ErrorKind as IOErrorKind;
use log::error;

mod logger;

use relayer::{run, db, config};

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    logger::init().map_err(|e| {
        // TODO: do we need to eprintln?
        eprintln!("Failed to initialize logger: {:?}", e);
        IOError::new(IOErrorKind::Other, format!("Logger init error: {:?}", e))
    })?;

    let db = db::init().await.map_err(|e| {
        error!("Failed to initialize database connection:");
        error!("  Error: {}", e);
        if let Some(source) = e.source() {
            error!("  Source: {}", source);
        }
        e
    })?;
    
    let contract_config = config::init().map_err(|e| {
        error!("Failed to initialize contract config:");
        error!("  Error: {}", e);
        if let Some(source) = e.source() {
            error!("  Source: {}", source);
        }
        e
    })?;

    run(&db, &contract_config).await.map_err(|e| {
        error!("Application execution failed:");
        error!("  Error: {}", e);
        if let Some(source) = e.source() {
            error!("  Source: {}", source);
        }
        e
    })?;

    Ok(())
}
