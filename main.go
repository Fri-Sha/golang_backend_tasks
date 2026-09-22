package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"main/handlers"
	"main/postgresql"
	"main/repository"
	"main/server"
	workerPackage "main/worker"

	"golang.org/x/net/context"
)

const port = "8080"

func main() {
	migrateDown, err := strconv.ParseBool(os.Getenv("DB_DOWN"))
	if err != nil {
		migrateDown = false
	}

	storage := postgresql.New(postgresql.DBconfig{
		Host:    os.Getenv("DB_HOST"),
		Port:    os.Getenv("DB_PORT"),
		User:    os.Getenv("DB_USER"),
		Pass:    os.Getenv("DB_PASSWORD"),
		DBName:  os.Getenv("DB_NAME"),
		SSLMode: "disable",
	}, migrateDown)

	defer storage.Close()

	worker := workerPackage.NewWorker(storage.DB)
	repos := repository.NewRepository(worker)
	handler := handlers.NewHandler(repos)

	ctx, cancelWorkers := context.WithCancel(context.Background())

	for i := 0; i < workerPackage.MaxWorkers; i++ {
		go worker.StartWorker(worker.TaskChan, worker.ExitChan, ctx)
	}

	srv := new(server.Server)
	go func() {
		if err := srv.Run(port, handler.InitRoute()); err != http.ErrServerClosed {
			log.Fatalf("Возникла ошибка при работе HTTP сервера: %s", err.Error())
		}
	}()

	log.Println("Сервер поднялся на порту " + port)

	log.Println("Проверка task со статусом new...")
	err, newCount := repos.CheckNewTasks()
	if err != nil {
		log.Fatalf("Возникла ошибка при проверке task со статусом new: %s", err.Error())
	} else {
		log.Println("Проверка task со статусом new прошла успешно, отправлено на обработку:", newCount)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Выключение сервера...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Возникла ошибка при выключении сервера: %s", err.Error())
	}

	log.Println("Выключение worker...")
	cancelWorkers()
	for i := 0; i < workerPackage.MaxWorkers; i++ {
		<-worker.ExitChan
	}
}
