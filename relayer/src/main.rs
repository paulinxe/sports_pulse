use std::error::Error;
use std::io::Error as IOError;
use std::io::ErrorKind as IOErrorKind;
use log::error;

use relayer::{run, config::db, config::contract, config::logger};
use relayer::services::blockchain_broadcaster::BlockchainBroadcaster;

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
    
    let contract_config = contract::init().map_err(|e| {
        error!("Failed to initialize contract config:");
        error!("  Error: {}", e);
        if let Some(source) = e.source() {
            error!("  Source: {}", source);
        }
        e
    })?;

    let broadcaster = BlockchainBroadcaster::new(contract_config);
    run(&db, &broadcaster).await.map_err(|e| {
        error!("Application execution failed:");
        error!("  Error: {}", e);
        if let Some(source) = e.source() {
            error!("  Source: {}", source);
        }
        e
    })?;

    Ok(())
}
