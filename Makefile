dev:
	@bash -c 'trap "kill 0" EXIT INT TERM; air & npm run watch:css & wait'

format:
	npx prettier --write ./web/templates --log-level warn
