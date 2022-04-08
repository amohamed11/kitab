build:
	go build --tags "fts5" -o bin/kitab main.go

serve: build
	./bin/kitab

run:
	go run --tags "fts5" main.go

clean:
	rm bin/*