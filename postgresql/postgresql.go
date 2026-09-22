package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Важно!
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBconfig struct {
	Host    string
	Port    string
	User    string
	Pass    string
	DBName  string
	SSLMode string
}
type Storage struct {
	DB *gorm.DB
}

var cfgDB DBconfig

const localHost = "localhost"
const localUser = "postgres"
const localDbname = "tasks"
const localPass = "postgres"
const localSslmode = "disable"
const localPort = "8092"

func New(cfg DBconfig, downMigration bool) *Storage {
	storagePath := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s port=%s",
		cfg.Host,
		cfg.User,
		cfg.DBName,
		cfg.Pass,
		cfg.SSLMode,
		cfg.Port)

	cfgDB = cfg

	gormDB, err := gorm.Open(postgres.Open(storagePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		// Повторная попытка подключения, возможно пытаемся поднять программу локально
		storagePath = fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s port=%s",
			localHost,
			localUser,
			localDbname,
			localPass,
			localSslmode,
			localPort)

		cfgDB = cfg

		gormDB, err = gorm.Open(postgres.Open(storagePath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Fatal("Ошибка подключения к базе данных:", err)
		}
	}

	sqlDB, _ := gormDB.DB()

	err = runMigrations(sqlDB, downMigration)
	if err != nil {
		log.Fatal(err)
	}

	return &Storage{DB: gormDB}
}

func (s *Storage) Close() {
	sqlDB, err := s.DB.DB()
	if err != nil {
		log.Fatal("Ошибка закрытия подключения к БД:", err)
	}

	_ = sqlDB.Close()
}

func runMigrations(db *sql.DB, downMigration bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Не удалось получить рабочую директорию")
	}
	log.Println("Рабочая директория:", wd)

	driver, err := postgresMigrate.WithInstance(db, &postgresMigrate.Config{
		DatabaseName: cfgDB.DBName,
		SchemaName:   "public",
	})
	if err != nil {
		return fmt.Errorf("Ошибка получения драйвера postgresMigrate: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		m, err = migrate.NewWithDatabaseInstance("file://../migrations", "postgres", driver)
		if err != nil {
			return fmt.Errorf("Ошибка создания инстанции миграции: %w", err)
		}
	}

	if downMigration {
		err = m.Down()
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("Ошибка при выполнении down миграций: %w", err)
		}
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Ошибка при выполнении миграций: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("Ошибка при получении версии миграций: %w", err)
	}

	log.Printf("Версия миграция базы данных: %d, dirty: %v", version, dirty)

	return nil
}
