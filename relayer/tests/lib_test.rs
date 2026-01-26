// Integration tests - these test the full application flow

use relayer::{run, config::db, Broadcaster};
use log::Level;
use mockall::predicate::*;
use mockall::mock;
use chrono::NaiveDateTime;
use uuid::Uuid;
mod in_memory_logger;
use in_memory_logger::InMemoryLogger;

// Create a mock for the Broadcaster trait
mock! {
    Broadcaster {}
    
    #[async_trait::async_trait]
    impl Broadcaster for Broadcaster {
        async fn broadcast(&self, m: &relayer::entity::match_entity::Match) -> Result<(), Box<dyn std::error::Error>>;
    }
}


#[tokio::test]
async fn test_run_successfully_ends_when_no_matches_are_found() {
    // Initialize the in-memory logger (ignore error if already initialized)
    let _ = InMemoryLogger::init(Level::Info);
    InMemoryLogger::clear();

    let db = db::init().await.unwrap();
    db.execute("TRUNCATE TABLE matches", &[]).await.unwrap();
    
    // Create a mock broadcaster that should not be called
    let mut mock_broadcaster = MockBroadcaster::new();
    mock_broadcaster
        .expect_broadcast()
        .times(0);

    let result = run(&db, &mock_broadcaster).await;
    assert!(result.is_ok(), "run() should handle empty database gracefully");

    // Check that the expected log message was outputted
    let logger = InMemoryLogger;
    let infos = logger.infos();
    let expected_message = "Found 0 signed matches to process";
    assert!(
        infos.iter().any(|msg| msg.contains(&expected_message)),
        "Expected to find log message: '{}'. Found info logs: {:?}",
        expected_message,
        infos
    );
}

#[tokio::test]
async fn test_run_broadcasts_match_result() {
    let _ = InMemoryLogger::init(Level::Info);
    InMemoryLogger::clear();

    let db = db::init().await.unwrap();
    db.execute("TRUNCATE TABLE matches", &[]).await.unwrap();
    
    let match_date = NaiveDateTime::parse_from_str("2025-12-19 00:00:00", "%Y-%m-%d %H:%M:%S")
        .expect("Failed to parse date");
    let match_end = match_date + chrono::Duration::hours(2);
    let match_id = Uuid::parse_str("e983312d-c32c-4902-8cb6-8147152fd476")
        .expect("Failed to parse UUID");
    
    db.execute(
        "INSERT INTO matches (id, canonical_id, home_team_id, away_team_id, home_team_score, away_team_score, signature, competition_id, start, \"end\", provider_match_id, provider, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
        &[
            &match_id,
            &"0x7ed54b4173481077ca259c17a51291beed5152f35d37e01142cd6ee2f771127f".to_string(),
            &1i32,
            &2i32,
            &1i32, // home_team_score
            &2i32, // away_team_score
            &"4f0fa54d6dd9629d5f1d6b0f17236f4f9f009b72be6e77bdc56a4d0d891c0c076f6c36472f7b667d5f63895424a19a19bc56f264e49699c58bb07ec0868440081c".to_string(),
            &1i32, // competition_id
            &match_date,
            &match_end,
            &"1234567890".to_string(), // provider_match_id
            &1i32, // provider
            &4i32, // status = 4 (signed)
        ],
    ).await.unwrap();

    // Create a mock broadcaster that expects broadcast to be called once with the correct Match structure
    let expected_canonical_id = "0x7ed54b4173481077ca259c17a51291beed5152f35d37e01142cd6ee2f771127f"
        .parse::<alloy::primitives::B256>()
        .expect("Failed to parse canonical_id");
    let expected_start_ymd = 20251219u32; // 2025-12-19 in YMD format
    let expected_signature = "4f0fa54d6dd9629d5f1d6b0f17236f4f9f009b72be6e77bdc56a4d0d891c0c076f6c36472f7b667d5f63895424a19a19bc56f264e49699c58bb07ec0868440081c"
        .parse::<alloy::primitives::Bytes>()
        .expect("Failed to parse signature");
    
    let mut mock_broadcaster = MockBroadcaster::new();
    mock_broadcaster
        .expect_broadcast()
        .times(1)
        .withf(move |actual_match: &relayer::entity::match_entity::Match| {
            actual_match.canonical_id == expected_canonical_id
                && actual_match.home_team_id == 1
                && actual_match.away_team_id == 2
                && actual_match.home_team_score == 1
                && actual_match.away_team_score == 2
                && actual_match.competition_id == 1
                && actual_match.start == expected_start_ymd
                && actual_match.id == match_id
                && actual_match.signature == expected_signature
        })
        .returning(|_| Ok(()));

    let result = run(&db, &mock_broadcaster).await;
    assert!(result.is_ok(), "run() should successfully broadcast the match");

    // Check that the expected log message was outputted
    let logger = InMemoryLogger;
    let infos = logger.infos();
    let expected_message = "Found 1 signed matches to process";
    assert!(
        infos.iter().any(|msg| msg.contains(&expected_message)),
        "Expected to find log message: '{}'. Found info logs: {:?}",
        expected_message,
        infos
    );
    
    let broadcast_message = "Broadcasted match: canonical_id=";
    assert!(
        infos.iter().any(|msg| msg.contains(&broadcast_message)),
        "Expected to find broadcast log message. Found info logs: {:?}",
        infos
    );
}
