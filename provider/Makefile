test:
	go test ./... 2>&1

run-football-org:
	go run ./cmd/provider football_org la_liga

run-apifootball:
	go run ./cmd/provider apifootball championship

check:
	golangci-lint run
