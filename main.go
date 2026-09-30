package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"main/handlers"
	"main/postgresql"
	"main/repository"
	"main/server"
	workerPackage "main/worker"

	"golang.org/x/net/context"
)

const programPort = "8080"
const localHost = "localhost"
const localUser = "postgres"
const localDbname = "tasks"
const localPass = "postgres"
const localPort = "8092"
const localDown = "false"

func main() {
	dbHost := getEnv("DB_HOST", localHost)
	dbPort := getEnv("DB_PORT", localPort)
	dbUser := getEnv("DB_USER", localUser)
	dbPass := getEnv("DB_PASSWORD", localPass)
	dbName := getEnv("DB_NAME", localDbname)
	dbDown := getEnv("DB_DOWN", localDown)

	migrateDown, err := strconv.ParseBool(dbDown)
	if err != nil {
		migrateDown = false
	}

	storage := postgresql.New(postgresql.DBconfig{
		Host:    dbHost,
		Port:    dbPort,
		User:    dbUser,
		Pass:    dbPass,
		DBName:  dbName,
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
		if err := srv.Run(programPort, handler.InitRoute()); err != http.ErrServerClosed {
			log.Fatalf("Возникла ошибка при работе HTTP сервера: %s", err.Error())
		}
	}()

	log.Println("Сервер поднялся на порту " + programPort)

	log.Println("Проверка task со статусом new...")
	err, newCount := repos.CheckNewTasks()
	if err != nil {
		log.Fatalf("Возникла ошибка при проверке task со статусом new: %s", err.Error())
	} else {
		log.Println("Проверка task со статусом new прошла успешно, отправлено на обработку:", newCount)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit

	log.Println("Получили сигнал:", sig)
	log.Println("Выключение сервера...")
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
	defer cancelTimeout()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Fatalf("Возникла ошибка при выключении сервера: %s", err.Error())
	}

	log.Println("Выключение worker...")
	cancelWorkers()
	for i := 0; i < workerPackage.MaxWorkers; i++ {
		<-worker.ExitChan
	}

	log.Println("Выключение программы...")
}

func getEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
