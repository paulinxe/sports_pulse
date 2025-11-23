sh-signer:
	docker-compose exec signer sh

generate-private-key:
	openssl ecparam -genkey -name secp256k1 -noout -out signer/private.key