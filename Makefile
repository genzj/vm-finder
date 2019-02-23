default: build

build:
	docker build -t vm-finder:latest . && \
	{ docker stop vm-finder >/dev/null 2>&1 || true ; } && \
	docker run --rm -d --name vm-finder vm-finder:latest sleep 3600 && \
	docker cp vm-finder:/app ./
