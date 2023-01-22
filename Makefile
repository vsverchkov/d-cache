build:
	go build -o bin/d-cache

run: build
	./bin/d-cache

runreplica: build
	./bin/d-cache --listenaddr :4000 --leaderaddr :3000