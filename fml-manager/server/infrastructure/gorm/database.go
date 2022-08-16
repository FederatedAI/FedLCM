// Copyright 2022 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gorm

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

const (
	dsnTemplate = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai"
)

// InitDB initialize the connection to the PostgresSql db
func InitDB() error {
	host := viper.GetString("postgres.host")
	port := viper.GetString("postgres.port")
	dbName := viper.GetString("postgres.db")
	user := viper.GetString("postgres.user")
	password := viper.GetString("postgres.password")
	if host == "" || port == "" || dbName == "" || user == "" || password == "" {
		panic("database information incomplete")
	}

	sslMode := viper.GetString("postgres.sslmode")
	if sslMode == "" {
		sslMode = "disable"
	}

	debugLog, _ := strconv.ParseBool(viper.GetString("postgres.debug"))
	loglevel := logger.Silent
	if debugLog {
		loglevel = logger.Info
	}

	dsn := fmt.Sprintf(dsnTemplate, host, port, user, password, dbName, sslMode)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      loglevel,
			Colorful:      false,
		},
	)
	if _db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger}); err != nil {
		return err
	} else {
		db = _db
	}
	return nil
}
