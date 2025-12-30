args := $(wordlist 2, 100, $(MAKECMDGOALS))

run:
	go run cmd/torrenter/main.go $(args)
