// Library crate - exposes the core functionality for testing
pub mod broadcast_result;
pub mod db;
pub mod config;

pub use broadcast_result::ContractConfig;

use std::error::Error;
use alloy::primitives::Bytes;

pub enum ErrorCodes {
    DatabaseConnectionError = 1,
    MissingEnvironmentVariable = 2,
    QueryExecutionError = 3,
}

const SIGNED_MATCH_STATUS: i32 = 4;

/// Core application logic - processes signed matches and broadcasts them
pub async fn run(
    db: &tokio_postgres::Client,
    contract_config: &ContractConfig,
) -> Result<(), Box<dyn Error>> {
    let query = "SELECT id, canonical_id, home_team_id, away_team_id, home_team_score, away_team_score, signature, competition_id, start FROM matches WHERE status = $1";
    
    println!("[INFO] Fetching signed matches from database: query='{}', status={}", query, SIGNED_MATCH_STATUS);
    
    let rows = db
        .query(query, &[&SIGNED_MATCH_STATUS])
        .await
        .map_err(|e| {
            eprintln!("[ERROR] Database query failed:");
            eprintln!("  Query: {}", query);
            eprintln!("  Status parameter: {}", SIGNED_MATCH_STATUS);
            eprintln!("  Error type: {:?}", e);
            eprintln!("  Error message: {}", e);
            eprintln!("  Error source: {:?}", e.source());
            Box::new(std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("Query execution error: query='{}', status={}, original_error={}", query, SIGNED_MATCH_STATUS, e)
            )) as Box<dyn Error>
        })?;
    
    println!("[INFO] Found {} signed matches to process", rows.len());

    for (idx, row) in rows.iter().enumerate() {
        let match_id_db: i32 = row.get("id");
        let canonical_id: String = row.get("canonical_id");
        
        let match_id = canonical_id
            .parse::<alloy::primitives::B256>()
            .map_err(|e| {
                eprintln!("[ERROR] Failed to parse canonical_id as B256:");
                eprintln!("  Match database ID: {}", match_id_db);
                eprintln!("  Canonical ID value: '{}'", canonical_id);
                eprintln!("  Canonical ID length: {} bytes", canonical_id.len());
                eprintln!("  Row index: {}/{}", idx + 1, rows.len());
                eprintln!("  Parse error: {}", e);
                Box::new(std::io::Error::new(
                    std::io::ErrorKind::InvalidData,
                    format!("Failed to parse canonical_id as B256 for match id={}, canonical_id='{}' (row {}/{}): {}", 
                        match_id_db, canonical_id, idx + 1, rows.len(), e)
                )) as Box<dyn Error>
            })?;

        let home_team_id: i32 = row.get("home_team_id");
        let away_team_id: i32 = row.get("away_team_id");
        let home_team_score: i16 = row.get("home_team_score");
        let away_team_score: i16 = row.get("away_team_score");
        let competition_id: i32 = row.get("competition_id");
        let match_date: i64 = row.get("start");
        
        let signature_str: String = row.get("signature");
        let signature: Bytes = signature_str
            .parse()
            .map_err(|e| {
                eprintln!("[ERROR] Failed to parse signature as Bytes:");
                eprintln!("  Match database ID: {}", match_id_db);
                eprintln!("  Canonical ID: '{}'", canonical_id);
                eprintln!("  Signature value: '{}'", signature_str);
                eprintln!("  Signature length: {} bytes", signature_str.len());
                eprintln!("  Row index: {}/{}", idx + 1, rows.len());
                eprintln!("  Parse error: {}", e);
                Box::new(std::io::Error::new(
                    std::io::ErrorKind::InvalidData,
                    format!("Failed to parse signature as Bytes for match id={}, canonical_id='{}', signature='{}' (row {}/{}): {}", 
                        match_id_db, canonical_id, signature_str, idx + 1, rows.len(), e)
                )) as Box<dyn Error>
            })?;
        
        println!("[INFO] Processing match: db_id={}, canonical_id='{}'", match_id_db, canonical_id);

        broadcast_result::broadcast(
            contract_config,
            match_id,
            competition_id as u32,
            home_team_id as u32,
            away_team_id as u32,
            home_team_score as u8,
            away_team_score as u8,
            match_date as u32,
            signature,
        ).await?;
    }

    Ok(())
}
