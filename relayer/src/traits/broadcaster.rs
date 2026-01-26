use std::error::Error;
use crate::entity::match_entity::Match;

#[async_trait::async_trait]
pub trait Broadcaster: Send + Sync {
    async fn broadcast(&self, m: &Match) -> Result<(), Box<dyn Error>>;
}