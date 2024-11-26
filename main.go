package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"zim-solar-inverter-mapper/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/lib/pq"
)

type IoTData struct {
	Device    string `json:"Device"`
	Timestamp string `json:"Timestamp"`
	ProVer    int    `json:"ProVer"`
	MinorVer  int    `json:"MinorVer"`
	SN        int64  `json:"SN"`
	Model     string `json:"model"`
	Status    Status `json:"Status"`
}

type Status struct {
	Tyield         float64 `json:"Tyield"`
	Dyield         float64 `json:"Dyield"`
	PF             float64 `json:"PF"`
	Pmax           int     `json:"Pmax"`
	Pac            int     `json:"Pac"`
	Sac            int     `json:"Sac"`
	Uab            int     `json:"Uab"`
	Ubc            int     `json:"Ubc"`
	Uca            int     `json:"Uca"`
	Ia             int     `json:"Ia"`
	Ib             int     `json:"Ib"`
	Ic             int     `json:"Ic"`
	Freq           int     `json:"Freq"`
	Tmod           float64 `json:"Tmod"`
	Tamb           float64 `json:"Tamb"`
	Mode           string  `json:"Mode"`
	Qac            int     `json:"Qac"`
	BusCapacitance float64 `json:"BusCapacitance"`
	AcCapacitance  float64 `json:"AcCapacitance"`
	Pdc            float64 `json:"Pdc"`
	PmaxLim        float64 `json:"PmaxLim"`
	SmaxLim        float64 `json:"SmaxLim"`
}

func main() {
	// 설정 가져오기
	mqttBroker := config.GetMQTTBroker()
	mqttTopic := config.GetMQTTTopic()
	//clientID := config.GetMQTTClientID()

	dbHost, dbPort, dbUser, dbPassword, dbName := config.GetPostgresConfig()

	// PostgreSQL 연결 문자열 구성
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	fmt.Println(psqlInfo)

	// 데이터베이스 연결
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("PostgreSQL 연결 실패: %v", err)
	}
	defer db.Close()

	// 연결 확인
	err = db.Ping()
	if err != nil {
		log.Fatalf("PostgreSQL 핑 실패: %v", err)
	}
	log.Println("PostgreSQL에 성공적으로 연결되었습니다.")

	// MQTT 클라이언트 옵션 설정
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttBroker)
	//opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("수신한 메시지: %s: %s\n", msg.Topic(), msg.Payload())

		var data IoTData
		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Printf("JSON 파싱 오류: %v", err)
			return
		}

		timestamp, err := time.Parse("2006-01-02 15:04:05", data.Timestamp)
		if err != nil {
			log.Printf("타임스탬프 파싱 오류: %v", err)
			return
		}

		// 데이터 인서트
		query := `
			INSERT INTO iot_data (
				device, timestamp, pro_ver, minor_ver, sn, model,
				tyield, dyield, pf, pmax, pac, sac,
				uab, ubc, uca, ia, ib, ic, freq,
				tmod, tamb, mode, qac, bus_capacitance,
				ac_capacitance, pdc, pmax_lim, smax_lim
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11, $12,
				$13, $14, $15, $16, $17, $18, $19,
				$20, $21, $22, $23, $24,
				$25, $26, $27, $28
			)
		`

		_, err = db.Exec(query,
			data.Device,
			timestamp,
			data.ProVer,
			data.MinorVer,
			data.SN,
			data.Model,
			data.Status.Tyield,
			data.Status.Dyield,
			data.Status.PF,
			data.Status.Pmax,
			data.Status.Pac,
			data.Status.Sac,
			data.Status.Uab,
			data.Status.Ubc,
			data.Status.Uca,
			data.Status.Ia,
			data.Status.Ib,
			data.Status.Ic,
			data.Status.Freq,
			data.Status.Tmod,
			data.Status.Tamb,
			data.Status.Mode,
			data.Status.Qac,
			data.Status.BusCapacitance,
			data.Status.AcCapacitance,
			data.Status.Pdc,
			data.Status.PmaxLim,
			data.Status.SmaxLim,
		)

		if err != nil {
			log.Printf("데이터베이스 삽입 오류: %v", err)
			return
		}

		log.Println("데이터베이스에 성공적으로 삽입되었습니다.")
	})

	// 클라이언트 생성 및 연결
	client := mqtt.NewClient(opts)
	fmt.Println("MQTT 연결중...", client.IsConnected())
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT 연결 실패: %v", token.Error())
	}
	log.Println("MQTT에 성공적으로 연결되었습니다.")

	// 토픽 구독
	if token := client.Subscribe(mqttTopic, 1, nil); token.Wait() && token.Error() != nil {
		log.Fatalf("토픽 구독 실패: %v", token.Error())
	}
	log.Printf("토픽 '%s'을(를) 구독 중입니다.\n", mqttTopic)

	// 애플리케이션이 종료되지 않도록 대기
	select {}
}
