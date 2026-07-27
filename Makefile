dev:
	@bash -c 'trap "kill 0" EXIT INT TERM; air & npm run watch:css & wait'

format:
	npx prettier --write ./web/templates --log-level warn
	gofmt -w .

test:
	@APP_BASE_DIR="/home/elias/Proyectos/poliplanner2/" go test ./... 2>&1 | sed -E '/no test files/d; s/(--- FAIL:.*|FAIL.*)/\x1b[31m&\x1b[0m/g; s/(--- PASS:.*|PASS.*)/\x1b[32m&\x1b[0m/g; s/(^ok\s+.*)/\x1b[32m&\x1b[0m/g'
