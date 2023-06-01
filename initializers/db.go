package initializers

import (
	"fmt"
	"log"
	"sync"

	"github.com/cploutarchou/go-kafka-rest/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	onceDB sync.Once  // guards DB initialization
	DB     *gorm.DB   // global connection to the DB
	dbMu   sync.Mutex // mutex for concurrent access to the DB
)

// ConnectDB connects to the database and performs migrations.
func ConnectDB(config *Config) {
	onceDB.Do(func() {
		var err error
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Athens",
			config.DBHost,
			config.DBUserName,
			config.DBUserPassword,
			config.DBName,
			config.DBPort,
		)
		dbMu.Lock()
		defer dbMu.Unlock()

		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to the Database!\n", err.Error())
		}

		DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
		DB.Logger = logger.Default.LogMode(logger.Info)

		log.Println("Running Migrations")
		err = DB.AutoMigrate(&models.User{})
		if err != nil {
			log.Fatal("Migration Failed:\n", err.Error())
		}

		log.Println("🚀 Successfully connected to the database!")
	})
}
