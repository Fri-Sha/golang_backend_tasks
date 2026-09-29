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

build-test:
	docker compose -f docker-compose.test.yml up --build --abort-on-container-exit

remove:
	docker rm full_db_postgres
	docker rm full_app
	docker rm full_db_test_postgres
	docker rm full_app_test