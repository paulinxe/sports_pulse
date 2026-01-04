use alloy::{
    primitives::{
        address,
        b256, Bytes,
    },
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
pub async fn broadcast() -> Result<(), Box<dyn Error>> {
    // https://alloy.rs/introduction/getting-started

    // This private key is NOT related with the Signer private key.
    // Using a hardcoded value for now from anvil node.
    let signer: PrivateKeySigner =
        "0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a".parse()?;
 
    // Instantiate a provider with the signer
    let provider = ProviderBuilder::new() 
        .wallet(signer) 
        .connect("http://oracle:8545") 
        .await?;
 
    // Setup MatchRegistry contract instance
    let match_registry = MatchRegistry::new(
        // TODO: get the address from the environment variable
        address!("0x5FC8d32690cc91D4c39d9d3abcBD16989F875707"),
        provider.clone()
    );

    let match_id = b256!("0x1234567890123456789012345678901234567890123456789012345678901234");
    let competition_id = 1u32;
    let home_team_id = 1u32;
    let away_team_id = 2u32;
    let home_team_score = 1u8;
    let away_team_score = 1u8;
    let match_date = 1u32;
    // Parse signature from hex string - can be any length (not just 32 bytes)
    // This properly handles Solidity's `bytes calldata` which is a dynamic byte array
    let signature: Bytes = "0x1234567890123456789012345678901234567890123456789012345678901234".parse()?;
 
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

