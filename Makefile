setup:
	git config core.hooksPath githooks
	go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema/
migrate:
	go run cmd/migrate/main.go up
rollback:
	go run cmd/migrate/main.go down
run:
	go run cmd/app/main.go