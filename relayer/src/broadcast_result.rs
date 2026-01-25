use alloy::{
    primitives::Bytes,
    providers::ProviderBuilder,
    signers::local::PrivateKeySigner,
    sol,
};
use std::error::Error;

pub struct ContractConfig {
    pub private_key: String,
    pub rpc: String,
    pub contract_address: alloy::primitives::Address,
}

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
pub async fn broadcast(
    config: &ContractConfig,
    match_id: alloy::primitives::B256,
    competition_id: u32,
    home_team_id: u32,
    away_team_id: u32,
    home_team_score: u8,
    away_team_score: u8,
    match_date: u32,
    signature: Bytes,
) -> Result<(), Box<dyn Error>> {
    // https://alloy.rs/introduction/getting-started

    // This private key is NOT related with the Signer private key.
    let signer: PrivateKeySigner = config.private_key.parse()?;
 
    // Instantiate a provider with the signer
    let provider = ProviderBuilder::new() 
        .wallet(signer) 
        .connect(&config.rpc) 
        .await?;

    // Setup MatchRegistry contract instance
    let match_registry = MatchRegistry::new(
        config.contract_address,
        provider.clone()
    );
 
    // Submit match result to the blockchain
    let tx = match_registry.submitMatch(
        match_id,
        competition_id,
        home_team_id,
        away_team_id,
        home_team_score,
        away_team_score,
        match_date,
        signature
    ).send().await?;

    let receipt = tx.get_receipt().await?;
    println!("Transaction receipt: {:?}", receipt);

    Ok(())
}

