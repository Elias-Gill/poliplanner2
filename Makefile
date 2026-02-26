dev:
	@bash -c 'trap "kill 0" EXIT INT TERM; air & npm run watch:css & wait'
