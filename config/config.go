package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type AppConfigProperties map[string]interface{}

var ConfInfo AppConfigProperties

// 초기화 함수
func init() {
	profile := os.Getenv("PROFILE")
	if profile == "" {
		profile = "dev" // 기본값 설정
	}

	if profile == "dev" {
		_, err := ReadConfigFile("config/config.json")
		if err != nil {
			log.Println("Failed to read config.json in dev mode:", err)
		}
	} else {
		ConfInfo = LoadConfigFromEnv()
	}
}

func ReadConfigFile(filename string) (AppConfigProperties, error) {
	ConfInfo = AppConfigProperties{}

	file, err := os.Open(filename)
	if err != nil {
		log.Println("Error opening config.json:", err)
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&ConfInfo); err != nil {
		log.Println("Error decoding config.json:", err)
		return nil, err
	}

	return ConfInfo, nil
}

func LoadConfigFromEnv() AppConfigProperties {
	conf := AppConfigProperties{}

	if broker := os.Getenv("MQTT_BROKER"); broker != "" {
		fmt.Println("broker: ", broker)
	}
	if topic := os.Getenv("MQTT_TOPIC"); topic != "" {
		conf["MQTTTopic"] = topic
	}
	if clientID := os.Getenv("MQTT_CLIENT_ID"); clientID != "" {
		conf["MQTTClientID"] = clientID
	}

	fmt.Println("MQTT Broker: ", conf["MQTTBroker"])
	fmt.Println("MQTT Topic: ", conf["MQTTTopic"])
	fmt.Println("MQTT Client ID: ", conf["MQTTClientID"])

	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		conf["DBHost"] = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		conf["DBPort"] = dbPort
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		conf["DBUser"] = dbUser
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		conf["DBPassword"] = dbPassword
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		conf["DBName"] = dbName
	}

	return conf
}

func GetAllowedOrigins() []string {
	if allowedOriginsInterface, exists := ConfInfo["AllowedOrigins"]; exists {
		switch v := allowedOriginsInterface.(type) {
		case string:
			return strings.Split(v, ",")
		case []interface{}:
			var origins []string
			for _, origin := range v {
				if strOrigin, ok := origin.(string); ok {
					origins = append(origins, strOrigin)
				}
			}
			return origins
		}
	}
	return nil
}

func GetMQTTBroker() string {
	if broker, exists := ConfInfo["MQTTBroker"]; exists {
		if b, ok := broker.(string); ok {
			return b
		}
	}
	return "tcp://localhost:1883" // 기본값
}

func GetMQTTTopic() string {
	if topic, exists := ConfInfo["MQTTTopic"]; exists {
		if t, ok := topic.(string); ok {
			return t
		}
	}
	return "iot/data" // 기본값
}

func GetMQTTClientID() string {
	if clientID, exists := ConfInfo["MQTTClientID"]; exists {
		if c, ok := clientID.(string); ok {
			return c
		}
	}
	return "edge-node-mapper" // 기본값
}

func GetPostgresConfig() (host string, port int, user, password, dbname string) {
	if hostVal, exists := ConfInfo["DBHost"]; exists {
		if h, ok := hostVal.(string); ok {
			host = h
		}
	} else {
		host = "localhost"
	}

	// DBPort
	if portVal, exists := ConfInfo["DBPort"]; exists {
		switch p := portVal.(type) {
		case float64:
			port = int(p)
		case int:
			port = p
		case string:
			parsedPort, err := strconv.Atoi(p)
			if err != nil {
				log.Printf("Invalid DBPort value '%s': %v. Using default port 5432.", p, err)
				port = 5432 // 기본값 설정
			} else {
				port = parsedPort
			}
		default:
			log.Printf("Unexpected type for DBPort: %T. Using default port 5432.", p)
			port = 5432
		}
	} else {
		port = 5432
	}

	if userVal, exists := ConfInfo["DBUser"]; exists {
		if u, ok := userVal.(string); ok {
			user = u
		}
	} else {
		user = "-"
	}

	if passwordVal, exists := ConfInfo["DBPassword"]; exists {
		if pw, ok := passwordVal.(string); ok {
			password = pw
		}
	} else {
		password = "-"
	}

	if dbnameVal, exists := ConfInfo["DBName"]; exists {
		if dn, ok := dbnameVal.(string); ok {
			dbname = dn
		}
	} else {
		dbname = "-"
	}

	return
}
