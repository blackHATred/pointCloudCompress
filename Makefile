gen-cert:
	openssl req -x509 -newkey rsa:2048 -nodes -keyout config/localhost-key.pem -out config/localhost.pem -subj "/CN=localhost" -days 365

run-benchmark:
	go run benchmark.go -dir=data -n=10 -voxel=0.1