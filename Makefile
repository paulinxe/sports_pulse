sh-provider:
	docker compose exec provider sh

sh-signer:
	docker compose exec signer sh

sh-relayer:
	docker compose exec relayer sh

sh-postgres:
	docker compose exec postgres sh

slither:
	docker compose run --rm slither
