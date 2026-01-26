use std::error::Error;
use alloy::primitives::Bytes;
use log::error;

pub struct Match {
    pub id: i32,
    pub canonical_id: alloy::primitives::B256,
    pub home_team_id: i32,
    pub away_team_id: i32,
    pub home_team_score: i16,
    pub away_team_score: i16,
    pub competition_id: i32,
    pub start: i64,
    pub signature: Bytes,
}

impl Match {
    pub fn from_db(row: &tokio_postgres::Row) -> Result<Self, Box<dyn Error>> {
        let canonical_id_str: String = row.get("canonical_id");
        let signature_str: String = row.get("signature");

        Ok(Self {
            id: row.get("id"),
            canonical_id: canonical_id_str.parse::<alloy::primitives::B256>()?,
            home_team_id: row.get("home_team_id"),
            away_team_id: row.get("away_team_id"),
            home_team_score: row.get("home_team_score"),
            away_team_score: row.get("away_team_score"),
            competition_id: row.get("competition_id"),
            start: row.get("start"),
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