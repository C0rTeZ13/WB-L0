setup:
	git config core.hooksPath githooks
	go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema/
migrate:
	docker exec l0-app-1 go run cmd/migrate/main.go up
rollback:
	docker exec l0-app-1 go run cmd/migrate/main.go down
run:
	docker exec l0-app-1 go run cmd/app/main.go
runTests:
	docker exec -e CONFIG_PATH=/app/config/local.yaml l0-app-1 go test ./...