package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io/ioutil"
	"strings"
	"time"
)

func initDatabase() {
	var err error
	var db *sql.DB

	host := viper.Get("host").(string)
	if host == "localhost" {
		host = ""
	}
	dbUrl = fmt.Sprintf("%s:%s@%s/%s",
		viper.Get("user").(string),
		viper.Get("password").(string),
		host,
		viper.Get("database").(string))

	fmtString := "CREATE TABLE IF NOT EXISTS %s " +
		" ENGINE=InnoDB DEFAULT CHARSET=utf8"
	tableCreateStmts := []string{
		`sensors (
    id smallint unsigned not null auto_increment UNIQUE,
    name CHAR(16) NOT NULL UNIQUE,
    primary key(id)
)`,
		`time_zones (
    id smallint unsigned not null auto_increment UNIQUE,
    name CHAR(16) NOT NULL UNIQUE,
    primary key(id)
)`,
		`readings (
    id int not null auto_increment unique,
    sensor_id smallint unsigned not null,
    temperature float not null,
    collected_at timestamp not null,
	zone_id smallint unsigned not null,
    primary key(id),
    constraint sensor_must_exist
        foreign key (sensor_id) references sensors (id)
        on delete cascade,
    constraint zone_must_exist
        foreign key (zone_id) references time_zones (id)
        on delete cascade,
    constraint no_repetition unique (sensor_id,
        collected_at, zone_id)
)`}

	if db, err = sql.Open("mysql", dbUrl); err != nil {
		log.Fatal().Msg("could not connect to database!")
		panic("could not connect to database!")
	}
	defer func() { _ = db.Close() }()

	for _, stmt := range tableCreateStmts {
		if _, err := db.Exec(fmt.Sprintf(fmtString, stmt)); err != nil {
			log.Fatal().Msg("could not create table!")
			panic("could not create table")
		}
	}
}

func getIdForSensor(db *sql.DB, id string) int {
	var err error
	var idNum int

	id = strings.ReplaceAll(
		strings.ReplaceAll(id, " ", ""),
		"\"",
		"")
	sensorQuery := fmt.Sprintf(
		"SELECT id FROM sensors WHERE name='%s'",
		id)
	if err = db.QueryRow(sensorQuery).Scan(&idNum); err != nil {
		sensorInsertStr := fmt.Sprintf(
			"INSERT INTO sensors(name) VALUES ('%s')",
			id)
		if _, err = db.Exec(sensorInsertStr); err != nil {
			log.Warn().Msgf("couldn't insert sensor record: %v",
			err)
			return -1
		}
	}
	log.Info().Msgf("inserted sensor %s into sensor table", id)
	if err = db.QueryRow(sensorQuery).Scan(&idNum); err != nil {
		log.Warn().Msgf(
			"could not read sensor %s, it should be there… %v",
			id, err)
		return -1
	}
	return idNum
}

func getZoneIdForSensor(db *sql.DB, timestamp time.Time) int {
	var err error
	var zoneNum int

	zoneName, _ := timestamp.Local().Zone()
	zoneName = strings.ReplaceAll(zoneName, "\"", "")
	if len(zoneName) > 16 {
		zoneName = zoneName[0:16]
	}
	zoneQuery := fmt.Sprintf(
		"SELECT id FROM time_zones WHERE name='%s'",
		zoneName)

	if err = db.QueryRow(zoneQuery).Scan(&zoneNum); err != nil {
		zoneInsertStr := fmt.Sprintf(
			"INSERT INTO time_zones(name) VALUES ('%s')",
			zoneName)
		if _, err = db.Exec(zoneInsertStr); err != nil {
			log.Warn().Msg("couldn't insert zone record!")
			return -1
		}
	}
	log.Info().Msgf("inserted zone %s into time zone table", zoneName)
	
	if nil != db.QueryRow(zoneQuery).Scan(&zoneNum) {
		log.Warn().Msgf(
			"could not read zone %s, it should be there…",
			zoneName)
		return -1
	}
	return zoneNum
}

func insertSensorRecord(db *sql.DB, sensor string, reading Reading) bool {
	var sensorId int
	var zoneId int
	var timestamp time.Time
	var mariaDbTimestamp string
	var err error

	if timestamp, err = time.Parse(time.RFC3339Nano,
		reading.Time); err != nil {
		log.Warn().Msg("error in parsing time in JSON")
		return false
	}
	mariaDbTimestamp = timestamp.Local().Format("2006-01-02 03:04:05")

	if sensorId = getIdForSensor(db, sensor); sensorId < 0 {
		return false
	}
	if zoneId = getZoneIdForSensor(db, timestamp); zoneId < 0 {
		return false
	}

	recordStr := fmt.Sprintf(
		"INSERT INTO readings(sensor_id, temperature, "+
			"collected_at, zone_id) "+
			"VALUES (%d, %3.2f, '%s', %d)",
		sensorId, reading.Temperature, mariaDbTimestamp, zoneId)
	if _, err = db.Exec(recordStr); err != nil {
		log.Warn().Msgf("could not insert %s / %3.2f C at %s: %v",
			strings.ReplaceAll(
				strings.ReplaceAll(sensor, " ", ""),
				"\"",
				""),
			reading.Temperature,
			mariaDbTimestamp,
			err)
		return false
	}
	log.Info().Msgf("inserted %s / %3.2f C at %s",
		sensor, reading.Temperature, mariaDbTimestamp)
	return true
}

func updateTable() {
	var err error
	var data []byte
	var db *sql.DB
	sensors := make(map[string]Reading)

	if data, err = ioutil.ReadFile("/tmp/kure.json"); err != nil {
		log.Warn().Msg("/tmp/kure.json could not be read")
		return
	}
	if err = json.Unmarshal(data, &sensors); err != nil {
		log.Warn().Msgf("bad JSON in sensor file: %v", err)
		return
	}

	for id := range sensors {
		if _, ok := sensorQueue[id]; !ok {
			sensorQueue[id] = make([]Reading, 0)
		}
		sensorQueue[id] = append(sensorQueue[id], sensors[id])
	}

	if db, err = sql.Open("mysql", dbUrl); err != nil {
		log.Warn().Msg("could not connect to database")
		return
	}
	defer func() { _ = db.Close() }()

	deleteList := make([]string, 0)
	for k := range sensorQueue {
		remnants := make([]Reading, 0)
		for _, value := range sensorQueue[k] {
			if !insertSensorRecord(db, k, value) {
				remnants = append(remnants, value)
			}
		}
		if len(remnants) > 0 {
			sensorQueue[k] = remnants
		} else {
			deleteList = append(deleteList, k)
		}
	}
	for _, keyName := range deleteList {
		delete(sensorQueue, keyName)
	}
}
