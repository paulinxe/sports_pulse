// Library crate - exposes the core functionality for testing
pub mod db;
pub mod config;
pub mod entity;
pub mod services;
pub use services::broadcast_result::broadcast;

use std::error::Error;
use log::{info, error, debug};
use crate::entity::match_entity::Match;
use crate::config::ContractConfig;

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
    debug!("Fetching signed matches from database");

    let query = "SELECT id, canonical_id, home_team_id, away_team_id, home_team_score, away_team_score, signature, competition_id, start FROM matches WHERE status = $1";
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

    for row in rows.iter() {
        let m = Match::from_db(&row)?;
        debug!("Processing match: db_id={}, canonical_id='{}'", m.id, m.canonical_id);

        broadcast(contract_config, &m).await?;
        info!("Broadcasted match: canonical_id='{}'", m.canonical_id);
    }

    Ok(())
}
