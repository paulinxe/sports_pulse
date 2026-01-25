// Integration tests - these test the full application flow
// Run with: cargo test --test integration_test

use relayer::{run, db::init, ContractConfig};

/// Helper function to create a test contract config
fn create_test_config() -> ContractConfig {
    ContractConfig {
        private_key: "0x0000000000000000000000000000000000000000000000000000000000000001".to_string(),
        rpc: "https://some-non-existent-rpc-url.com".to_string(),
        contract_address: alloy::primitives::Address::ZERO,
    }
}

#[tokio::test]
async fn test_run_successfully_ends_when_no_matches_are_found() {
    let db = init().await.unwrap();
    db.execute("TRUNCATE TABLE matches", &[]).await.unwrap();
    let config = create_test_config();

    let result = run(&db, &config).await;
    assert!(result.is_ok(), "run() should handle empty database gracefully");
}
