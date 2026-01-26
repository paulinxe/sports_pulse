use std::error::Error;
use alloy::primitives::Bytes;
use log::error;
use uuid::Uuid;
use chrono::{NaiveDateTime, Datelike};

pub struct Match {
    pub id: Uuid,
    pub canonical_id: alloy::primitives::B256,
    pub home_team_id: i32,
    pub away_team_id: i32,
    pub home_team_score: i32,
    pub away_team_score: i32,
    pub competition_id: i32,
    pub start: u32, // YMD format: YYYYMMDD (e.g., 20250126 for 2025-01-26)
    pub signature: Bytes,
}

impl Match {
    pub fn from_db(row: &tokio_postgres::Row) -> Result<Self, Box<dyn Error>> {
        let canonical_id_str: String = row.get("canonical_id");
        let signature_str: String = row.get("signature");
        let start_timestamp: NaiveDateTime = row.get("start");
        // Convert timestamp to YMD format: YYYYMMDD (e.g., 20250126 for 2025-01-26)
        let start_ymd = (start_timestamp.year() as u32 * 10000)
            + (start_timestamp.month() * 100)
            + (start_timestamp.day() as u32);

        Ok(Self {
            id: row.get("id"),
            canonical_id: canonical_id_str.parse::<alloy::primitives::B256>()?,
            home_team_id: row.get("home_team_id"),
            away_team_id: row.get("away_team_id"),
            home_team_score: row.get("home_team_score"),
            away_team_score: row.get("away_team_score"),
            competition_id: row.get("competition_id"),
            start: start_ymd,
            signature: signature_str
            .parse()
            .map_err(|e| {
                error!("Failed to parse signature as Bytes:");
                error!("  Canonical ID: '{}'", canonical_id_str);
                error!("  Signature value: '{}'", signature_str);
                error!("  Parse error: {}", e);
                Box::new(e) as Box<dyn Error>
            })?,
        })
    }
}