run:
	docker compose up

sh-provider:
	docker-compose exec provider sh

sh-signer:
	docker-compose exec signer sh

generate-private-key:
	openssl ecparam -genkey -name secp256k1 -noout -out signer/private.key

sh-postgres:
	docker exec -it fan_token_pulse-postgres-1 /bin/sh