build:
	docker compose up --build

start:
	docker start full_db_postgres
	docker start full_app

stop:
	docker stop full_db_postgres
	docker stop full_app

start_db:
	docker start full_db_postgres

stop_db:
	docker stop full_db_postgres

start_app:
	docker start full_app

stop_app:
	docker stop full_app

remove:
	docker rm full_db_postgres
	docker rm full_app