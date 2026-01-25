// Library crate - exposes the core functionality for testing
pub mod broadcast_result;
pub mod db;
pub mod config;

pub use broadcast_result::ContractConfig;

use std::error::Error;
use alloy::primitives::Bytes;
use log::{info, error};

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
    
    info!("Fetching signed matches from database: query='{}', status={}", query, SIGNED_MATCH_STATUS);
    
    let rows = db
        .query(query, &[&SIGNED_MATCH_STATUS])
        .await
        .map_err(|e| {
            error!("Database query failed:");
            error!("  Query: {}", query);
            error!("  Status parameter: {}", SIGNED_MATCH_STATUS);
            error!("  Error type: {:?}", e);
            error!("  Error message: {}", e);
            error!("  Error source: {:?}", e.source());
            Box::new(e) as Box<dyn Error>
        })?;
    
    info!("Found {} signed matches to process", rows.len());

    for (idx, row) in rows.iter().enumerate() {
        let match_id_db: i32 = row.get("id");
        let canonical_id_db: String = row.get("canonical_id");
        
        let canonical_id = canonical_id_db
            .parse::<alloy::primitives::B256>()
            .map_err(|e| {
                error!("Failed to parse canonical_id as B256:");
                error!("  Match database ID: {}", match_id_db);
                error!("  Canonical ID value: '{}'", canonical_id_db);
                error!("  Canonical ID length: {} bytes", canonical_id_db.len());
                error!("  Row index: {}/{}", idx + 1, rows.len());
                error!("  Parse error: {}", e);
                Box::new(e) as Box<dyn Error>
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
                error!("Failed to parse signature as Bytes:");
                error!("  Match database ID: {}", match_id_db);
                error!("  Canonical ID: '{}'", canonical_id_db);
                error!("  Signature value: '{}'", signature_str);
                error!("  Signature length: {} bytes", signature_str.len());
                error!("  Row index: {}/{}", idx + 1, rows.len());
                error!("  Parse error: {}", e);
                Box::new(e) as Box<dyn Error>
            })?;
        
        info!("Processing match: db_id={}, canonical_id='{}'", match_id_db, canonical_id_db);

        broadcast_result::broadcast(
            contract_config,
            canonical_id,
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
