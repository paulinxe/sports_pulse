use crate::config::contract::ContractConfig;
use crate::traits::broadcaster::Broadcaster;
use crate::entity::match_entity::Match;
use alloy::{
    providers::ProviderBuilder,
    signers::local::PrivateKeySigner,
    sol,
};
use std::error::Error;
sol! { 
    #[sol(rpc)] 
    contract MatchRegistry { 
        function submitMatch(
            bytes32 matchId,
            uint32 competitionId,
            uint32 homeTeamId,
            uint32 awayTeamId,
            uint8 homeTeamScore,
            uint8 awayTeamScore,
            uint32 matchDate,
            bytes calldata signature
        ) external;
    }
}

/// Production implementation that broadcasts to the blockchain
pub struct BlockchainBroadcaster {
    config: ContractConfig,
}

impl BlockchainBroadcaster {
    pub fn new(config: ContractConfig) -> Self {
        Self { config }
    }
}

#[async_trait::async_trait]
impl Broadcaster for BlockchainBroadcaster {
    async fn broadcast(&self, m: &Match) -> Result<(), Box<dyn Error>> {
        // https://alloy.rs/introduction/getting-started

        // This private key is NOT related with the Signer private key.
        let signer: PrivateKeySigner = self.config.private_key.parse()?;
     
        // Instantiate a provider with the signer
        let provider = ProviderBuilder::new() 
            .wallet(signer) 
            .connect(&self.config.rpc) 
            .await?;

        // Setup MatchRegistry contract instance
        let match_registry = MatchRegistry::new(
            self.config.contract_address,
            provider.clone()
        );
     
        // Submit match result to the blockchain
        let tx = match_registry.submitMatch(
            m.canonical_id,
            m.competition_id as u32,
            m.home_team_id as u32,
            m.away_team_id as u32,
            m.home_team_score as u8,
            m.away_team_score as u8,
            m.start,
            m.signature.clone() // TODO: try to find a better way
        ).send().await?;

        let receipt = tx.get_receipt().await?;
        println!("Transaction receipt: {:?}", receipt);

        Ok(())
    }
}