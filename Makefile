NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
BLUE_COLOR=\033[94;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

test:
	@echo "$(OK_COLOR)==> Testing$(NO_COLOR)"
	@DATABASE_URL=postgres://localhost/chatbus go test -race -cover ./server/...
	@echo "$(OK_COLOR)==> Testing Complete!$(NO_COLOR)"

run:
	@echo "$(OK_COLOR)==> Running$(NO_COLOR)"
	@DATABASE_URL=postgres://localhost/chatbus go run main.go

db:
	@echo "$(OK_COLOR)==> Wiping DB$(NO_COLOR)"
	@dropdb chatbus --if-exists;
	@echo "DROP DATABASE"
	@createdb chatbus;
	@echo "CREATE DATABASE"
	@echo "$(OK_COLOR)==> Wiping DB Done!$(NO_COLOR)"

cover:
	@echo "$(OK_COLOR)==> Coverage$(NO_COLOR)"
	@DATABASE_URL=postgres://localhost/chatbus ROOT="./" sh ./_util/coverage.sh
	@echo "$(OK_COLOR)==> Coverage Complete!$(NO_COLOR)"
