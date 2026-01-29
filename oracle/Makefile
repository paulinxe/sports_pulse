test:
	forge test

test-one:
	forge test -vvv --match-test $(test)

deploy:
	@if [ -z "$(RPC_URL)" ] || [ -z "$(DEPLOYER_PRIVATE_KEY)" ] || [ -z "$(AUTHORIZED_SIGNER_ADDRESS)" ] || [ -z "$(CONTRACTS_OWNER_ADDRESS)" ]; then \
		echo "Error: RPC_URL, DEPLOYER_PRIVATE_KEY, AUTHORIZED_SIGNER_ADDRESS, and CONTRACTS_OWNER_ADDRESS must be set"; \
		echo "Usage: make deploy RPC_URL=<url> DEPLOYER_PRIVATE_KEY=<key> AUTHORIZED_SIGNER_ADDRESS=<address> CONTRACTS_OWNER_ADDRESS=<address>"; \
		exit 1; \
	fi
	forge script script/Deploy.s.sol:Deploy --rpc-url $(RPC_URL) --broadcast