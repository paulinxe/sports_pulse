use alloy::{
    network::TransactionBuilder,
    primitives::{
        address,
        utils::Unit,
        U256,
    },
    providers::{Provider, ProviderBuilder},
    rpc::types::TransactionRequest,
    signers::local::PrivateKeySigner,
};
use std::error::Error;

pub async fn test_connection() -> Result<(), Box<dyn Error>> {
    // https://alloy.rs/introduction/getting-started

    // Initialize a signer with a private key
    let signer: PrivateKeySigner =
        "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80".parse()?;
 
    // Instantiate a provider with the signer and a local anvil node
    let provider = ProviderBuilder::new() 
        .wallet(signer) 
        .connect("http://oracle:8545") 
        .await?;
 
    // Prepare a transaction request to send 100 ETH to Alice
    let alice = address!("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"); 
    let value = Unit::ETHER.wei().saturating_mul(U256::from(100)); 
    let tx = TransactionRequest::default() 
        .with_to(alice) 
        .with_value(value); 
 
    // Send the transaction and wait for the broadcast
    let pending_tx = provider.send_transaction(tx).await?; 
    println!("Pending transaction... {}", pending_tx.tx_hash());
 
    // Wait for the transaction to be included and get the receipt
    let receipt = pending_tx.get_receipt().await?; 
    println!(
        "Transaction included in block {}",
        receipt.block_number.expect("Failed to get block number")
    );

    Ok(())
}

