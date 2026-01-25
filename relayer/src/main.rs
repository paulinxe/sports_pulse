use std::error::Error;

use relayer::{run, db, config};

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let db = db::init().await.map_err(|e| {
        eprintln!("[ERROR] Failed to initialize database connection:");
        eprintln!("  Error: {}", e);
        if let Some(source) = e.source() {
            eprintln!("  Source: {}", source);
        }
        e
    })?;
    
    let contract_config = config::init().map_err(|e| {
        eprintln!("[ERROR] Failed to initialize contract config:");
        eprintln!("  Error: {}", e);
        if let Some(source) = e.source() {
            eprintln!("  Source: {}", source);
        }
        e
    })?;

    run(&db, &contract_config).await.map_err(|e| {
        eprintln!("[ERROR] Application execution failed:");
        eprintln!("  Error: {}", e);
        if let Some(source) = e.source() {
            eprintln!("  Source: {}", source);
        }
        e
    })?;

    Ok(())
}
