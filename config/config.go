// config/config.go

package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// AppConfigProperties는 설정을 저장하는 맵입니다.
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

// ReadConfigFile은 JSON 파일에서 설정을 읽어 ConfInfo에 저장합니다.
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

// LoadConfigFromEnv는 프로덕션 모드에서 환경 변수로부터 설정을 로드합니다.
func LoadConfigFromEnv() AppConfigProperties {
	conf := AppConfigProperties{}

	// 예시: MQTT 설정
	if broker := os.Getenv("MQTT_BROKER"); broker != "" {
		conf["MQTTBroker"] = broker
	}
	if topic := os.Getenv("MQTT_TOPIC"); topic != "" {
		conf["MQTTTopic"] = topic
	}
	if clientID := os.Getenv("MQTT_CLIENT_ID"); clientID != "" {
		conf["MQTTClientID"] = clientID
	}

	// PostgreSQL 설정
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

	// 기타 설정 추가 가능

	return conf
}

// GetAllowedOrigins는 설정에서 AllowedOrigins를 가져옵니다.
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

// GetMQTTBroker는 설정에서 MQTT 브로커 주소를 가져옵니다.
func GetMQTTBroker() string {
	if broker, exists := ConfInfo["MQTTBroker"]; exists {
		if b, ok := broker.(string); ok {
			return b
		}
	}
	return "tcp://localhost:1883" // 기본값
}

// GetMQTTTopic는 설정에서 MQTT 토픽을 가져옵니다.
func GetMQTTTopic() string {
	if topic, exists := ConfInfo["MQTTTopic"]; exists {
		if t, ok := topic.(string); ok {
			return t
		}
	}
	return "iot/data" // 기본값
}

// GetMQTTClientID는 설정에서 MQTT 클라이언트 ID를 가져옵니다.
func GetMQTTClientID() string {
	if clientID, exists := ConfInfo["MQTTClientID"]; exists {
		if c, ok := clientID.(string); ok {
			return c
		}
	}
	return "edge-node-mapper" // 기본값
}

// GetPostgresConfig는 PostgreSQL 설정을 가져옵니다.
func GetPostgresConfig() (host string, port int, user, password, dbname string) {
	if hostVal, exists := ConfInfo["DBHost"]; exists {
		if h, ok := hostVal.(string); ok {
			host = h
		}
	} else {
		host = "localhost"
	}

	if portVal, exists := ConfInfo["DBPort"]; exists {
		// JSON에서는 숫자가 float64로 디코딩될 수 있음
		switch p := portVal.(type) {
		case float64:
			port = int(p)
		case int:
			port = p
		}
	} else {
		port = 5432
	}

	if userVal, exists := ConfInfo["DBUser"]; exists {
		if u, ok := userVal.(string); ok {
			user = u
		}
	} else {
		user = "yourusername"
	}

	if passwordVal, exists := ConfInfo["DBPassword"]; exists {
		if pw, ok := passwordVal.(string); ok {
			password = pw
		}
	} else {
		password = "yourpassword"
	}

	if dbnameVal, exists := ConfInfo["DBName"]; exists {
		if dn, ok := dbnameVal.(string); ok {
			dbname = dn
		}
	} else {
		dbname = "yourdbname"
	}

	return
}
