run:
	docker compose up postgres signer provider mock-api

sh-provider:
	docker compose exec provider sh

sh-signer:
	docker compose exec signer sh

sh-relayer:
	docker compose exec relayer sh

sh-postgres:
	docker compose exec postgres sh

generate-private-key:
	openssl ecparam -genkey -name secp256k1 -noout -out signer/private.key

slither:
	docker compose run --rm slither