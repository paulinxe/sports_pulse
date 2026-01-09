test:
	forge test

test-one:
	forge test -vvv --match-test $(test)

deploy:
	@if [ -z "$(DEPLOYER_PRIVATE_KEY)" ] || [ -z "$(AUTHORIZED_SIGNER_ADDRESS)" ] || [ -z "$(CONTRACTS_OWNER_ADDRESS)" ]; then \
		echo "Error: DEPLOYER_PRIVATE_KEY, AUTHORIZED_SIGNER_ADDRESS, and CONTRACTS_OWNER_ADDRESS must be set"; \
		exit 1; \
	fi
	forge script script/Deploy.s.sol:Deploy --broadcast