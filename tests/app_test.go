package tests

import (
	"log"
	"os"
	"strconv"
	"testing"

	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"main/handlers"
	"main/models"
	"main/postgresql"
	"main/repository"
	"main/server"
	workerPackage "main/worker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const programPort = "8080"
const localHost = "localhost"
const localUser = "postgres"
const localDbname = "tasks_test"
const localPass = "postgres"
const localPort = "8093"
const localDown = "true"

func TestValidateTask(t *testing.T) {
	var task models.Tasks

	task = models.Tasks{
		Id:        0,
		Title:     "Test Task",
		Status:    models.StatusNew.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repository.ValidateTask(task, false)
	log.Println("Task 1 результат проверки:", err, ", ожидалось:", repository.ErrInvalidTaskId)
	require.ErrorIs(t, err, repository.ErrInvalidTaskId)

	// task = models.Tasks{
	// 	Id:        1,
	// 	Title:     "Test Task",
	// 	Status:    models.StatusFailed.String(),
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// }
	// err = repository.ValidateTask(task, false)
	// log.Println("Task 2 результат проверки:", err, ", ожидалось:", ErrTaskFailed)
	// require.ErrorIs(t, err, repository.ErrTaskFailed)

	task = models.Tasks{
		Id:        1,
		Title:     "Test Task",
		Status:    models.StatusNew.String(),
		CreatedAt: time.Now(),
	}
	err = repository.ValidateTask(task, false)
	log.Println("Task 2 результат проверки:", err, ", ожидалось:", repository.ErrTaskNoTimestamps)
	require.ErrorIs(t, err, repository.ErrTaskNoTimestamps)

	task = models.Tasks{
		Id:        1,
		Title:     "Test Task",
		Status:    models.StatusNew.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repository.ValidateTask(task, false)
	log.Println("Task 3 результат проверки:", err, ", ошибки не ожидалось")
	require.ErrorIs(t, err, nil)
}

func TestServer(t *testing.T) {
	dbHost := getEnv("TEST_DB_HOST", localHost)
	dbPort := getEnv("TEST_DB_PORT", localPort)
	dbUser := getEnv("TEST_DB_USER", localUser)
	dbPass := getEnv("TEST_DB_PASSWORD", localPass)
	dbName := getEnv("TEST_DB_NAME", localDbname)
	dbDown := getEnv("TEST_DB_DOWN", localDown)

	migrateDown, err := strconv.ParseBool(dbDown)
	if err != nil {
		migrateDown = true
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
	router := handler.InitRoute()

	log.Println("Запуск сервера...")
	srv := new(server.Server)
	go func() {
		if err := srv.Run(programPort, handler.InitRoute()); err != http.ErrServerClosed {
			log.Fatalf("Возникла ошибка при работе HTTP сервера: %s", err.Error())
		}
	}()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(res, req)

	equal := assert.Equal(t, http.StatusOK, res.Code)
	if !equal {
		log.Println("Health-check failed. Response code:", res.Code)
	} else {
		log.Println("Health-check passed. Response code:", res.Code)
	}

	log.Println("Выключение сервера...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Возникла ошибка при выключении сервера: %s", err.Error())
	}
}

func TestWorkers(t *testing.T) {
	var task models.Tasks

	dbHost := getEnv("TEST_DB_HOST", localHost)
	dbPort := getEnv("TEST_DB_PORT", localPort)
	dbUser := getEnv("TEST_DB_USER", localUser)
	dbPass := getEnv("TEST_DB_PASSWORD", localPass)
	dbName := getEnv("TEST_DB_NAME", localDbname)
	dbDown := getEnv("TEST_DB_DOWN", localDown)

	migrateDown, err := strconv.ParseBool(dbDown)
	if err != nil {
		migrateDown = true
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
	router := handler.InitRoute()

	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < workerPackage.MaxWorkers; i++ {
		go worker.StartWorker(worker.TaskChan, worker.ExitChan, ctx)
	}

	models.IsTaskInProcess = make(map[uint64]bool)

	log.Println("Запуск сервера...")
	srv := new(server.Server)
	go func() {
		if err := srv.Run(programPort, handler.InitRoute()); err != http.ErrServerClosed {
			log.Fatalf("Возникла ошибка при работе HTTP сервера: %s", err.Error())
		}
	}()

	task.Title = "test 1"
	taskJson, _ := json.Marshal(task)
	req, _ := http.NewRequest(http.MethodPost, "/tasks", strings.NewReader(string(taskJson)))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	_ = json.Unmarshal(res.Body.Bytes(), &task)
	taskId1 := task.Id
	equal := assert.Equal(t, http.StatusCreated, res.Code)
	if !equal {
		log.Println("Ошибка создания task"+strconv.FormatUint(taskId1, 10)+". Response code:", res.Code)
	} else {
		log.Println("Task "+strconv.FormatUint(taskId1, 10)+" успешно создан. Response code:", res.Code)
	}

	task.Title = "test 2"
	taskJson, _ = json.Marshal(task)
	req, _ = http.NewRequest(http.MethodPost, "/tasks", strings.NewReader(string(taskJson)))
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	_ = json.Unmarshal(res.Body.Bytes(), &task)
	taskId2 := task.Id
	equal = assert.Equal(t, http.StatusCreated, res.Code)
	if !equal {
		log.Println("Ошибка создания task"+strconv.FormatUint(taskId2, 10)+". Response code:", res.Code)
	} else {
		log.Println("Task "+strconv.FormatUint(taskId2, 10)+" успешно создан. Response code:", res.Code)
	}

	task.Title = "fail 3"
	taskJson, _ = json.Marshal(task)
	req, _ = http.NewRequest(http.MethodPost, "/tasks", strings.NewReader(string(taskJson)))
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	_ = json.Unmarshal(res.Body.Bytes(), &task)
	taskId3 := task.Id
	equal = assert.Equal(t, http.StatusCreated, res.Code)
	if !equal {
		log.Println("Ошибка создания task "+strconv.FormatUint(taskId3, 10)+". Response code:", res.Code)
	} else {
		log.Println("Task "+strconv.FormatUint(taskId3, 10)+" успешно создан. Response code:", res.Code)
	}

	log.Println("Выключение сервера...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Возникла ошибка при выключении сервера: %s", err.Error())
	}

	log.Println("Выключение worker...")
	cancel()
	for i := 0; i < workerPackage.MaxWorkers; i++ {
		<-worker.ExitChan
	}

	req, _ = http.NewRequest(http.MethodGet, "/tasks/"+strconv.FormatUint(taskId1, 10), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	_ = json.Unmarshal(res.Body.Bytes(), &task)
	equal = assert.Equal(t, task.Status, models.StatusDone.String())
	if !equal {
		log.Println("Ошибка статуса task "+strconv.FormatUint(taskId1, 10)+". Ожидалось: "+models.StatusDone.String()+", получено:", task.Status)
	} else {
		log.Println("Task "+strconv.FormatUint(taskId1, 10)+" получил правильный статус. Ожидалось: "+models.StatusDone.String()+", получено:", task.Status)
	}

	req, _ = http.NewRequest(http.MethodGet, "/tasks/"+strconv.FormatUint(taskId2, 10), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	_ = json.Unmarshal(res.Body.Bytes(), &task)
	equal = assert.Equal(t, task.Status, models.StatusDone.String())
	if !equal {
		log.Println("Ошибка статуса task "+strconv.FormatUint(taskId2, 10)+". Ожидалось: "+models.StatusDone.String()+", получено:", task.Status)
	} else {
		log.Println("Task "+strconv.FormatUint(taskId2, 10)+" получил правильный статус. Ожидалось: "+models.StatusDone.String()+", получено:", task.Status)
	}

	req, _ = http.NewRequest(http.MethodGet, "/tasks/"+strconv.FormatUint(taskId3, 10), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	_ = json.Unmarshal(res.Body.Bytes(), &task)
	equal = assert.Equal(t, task.Status, models.StatusFailed.String())
	if !equal {
		log.Println("Ошибка статуса task "+strconv.FormatUint(taskId3, 10)+". Ожидалось: "+models.StatusDone.String()+", получено:", task.Status)
	} else {
		log.Println("Task "+strconv.FormatUint(taskId3, 10)+" получил правильный статус. Ожидалось: "+models.StatusFailed.String()+", получено:", task.Status)
	}
}

func getEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
