use std::error::Error;

pub fn init() -> Result<crate::broadcast_result::ContractConfig, Box<dyn Error>> {
    let private_key = std::env::var("RELAYER_PRIVATE_KEY").unwrap_or_else(|_| {
        eprintln!("Error: RELAYER_PRIVATE_KEY environment variable is not set");
        std::process::exit(crate::ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let rpc = std::env::var("RPC_URL").unwrap_or_else(|_| {
        eprintln!("Error: RPC_URL environment variable is not set");
        std::process::exit(crate::ErrorCodes::MissingEnvironmentVariable as i32);
    });

    let contract_address = std::env::var("ORACLE_CONTRACT_ADDRESS").unwrap_or_else(|_| {
        eprintln!("Error: ORACLE_CONTRACT_ADDRESS environment variable is not set");
        std::process::exit(crate::ErrorCodes::MissingEnvironmentVariable as i32);
    }).parse::<alloy::primitives::Address>()
        .map_err(|e| {
            eprintln!("Error parsing ORACLE_CONTRACT_ADDRESS: {}", e);
            std::process::exit(crate::ErrorCodes::MissingEnvironmentVariable as i32);
        })?;

    Ok(crate::broadcast_result::ContractConfig {
        private_key,
        rpc,
        contract_address,
    })
}
