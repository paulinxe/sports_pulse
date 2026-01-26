// Integration tests - these test the full application flow

use relayer::{run, config::db, config::contract::ContractConfig};
use log::Level;
mod in_memory_logger;
use in_memory_logger::InMemoryLogger;

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
    // Initialize the in-memory logger
    let logger = InMemoryLogger::init(Level::Info).unwrap();
    InMemoryLogger::clear();

    let db = db::init().await.unwrap();
    db.execute("TRUNCATE TABLE matches", &[]).await.unwrap();
    let config = create_test_config();

    let result = run(&db, &config).await;
    assert!(result.is_ok(), "run() should handle empty database gracefully");

    // Check that the expected log message was outputted
    let infos = logger.infos();
    let expected_message = "Found 0 signed matches to process";
    assert!(
        infos.iter().any(|msg| msg.contains(&expected_message)),
        "Expected to find log message: '{}'. Found info logs: {:?}",
        expected_message,
        infos
    );
}
