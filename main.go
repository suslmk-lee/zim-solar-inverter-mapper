// main.go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	Pmax           float64 `json:"Pmax"`
	Pac            int     `json:"Pac"`
	Sac            float64 `json:"Sac"`
	Uab            float64 `json:"Uab"`
	Ubc            float64 `json:"Ubc"`
	Uca            float64 `json:"Uca"`
	Ia             float64 `json:"Ia"`
	Ib             float64 `json:"Ib"`
	Ic             float64 `json:"Ic"`
	Freq           float64 `json:"Freq"`
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
	// 로그 설정
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("애플리케이션 시작")

	// 설정 가져오기
	mqttBroker := config.GetMQTTBroker()
	mqttTopic := config.GetMQTTTopic()
	clientID := config.GetMQTTClientID()

	// PostgreSQL DSN 가져오기
	psqlInfo := config.GetPostgresDSN()

	// 설정 값 로그 (비밀번호는 제외)
	log.Printf("MQTT Broker: %s", mqttBroker)
	log.Printf("MQTT Topic: %s", mqttTopic)
	log.Printf("MQTT Client ID: %s", clientID)
	log.Println("PostgreSQL DSN: [REDACTED]") // 보안을 위해 실제 DSN은 출력하지 않음

	// 데이터베이스 연결
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("PostgreSQL 연결 실패: %v", err)
	}
	defer db.Close()

	log.Println("PostgreSQL 연결 시도 중...")

	// 연결 확인
	err = db.Ping()
	if err != nil {
		log.Fatalf("PostgreSQL 핑 실패: %v", err)
	}
	log.Println("PostgreSQL에 성공적으로 연결되었습니다.")

	// MQTT 클라이언트 옵션 설정
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", mqttBroker))
	opts.SetClientID(clientID)
	opts.SetCleanSession(false) // 세션 유지
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetKeepAlive(60 * time.Second) // KeepAlive 설정
	opts.SetProtocolVersion(4)          // MQTT v3.1.1

	// 로그 레벨 설정 (옵션)
	// opts.SetLogger(&mqttLogger{})

	// 연결 신호 채널 생성
	connectedChan := make(chan struct{})

	// 커넥션 핸들러 추가
	opts.OnConnect = func(client mqtt.Client) {
		log.Println("MQTT에 성공적으로 연결되었습니다.")
		// 연결 신호 전송
		select {
		case connectedChan <- struct{}{}:
		default:
		}

		// 토픽 구독 시도
		if token := client.Subscribe(mqttTopic, 1, messageHandler(db)); token.Wait() && token.Error() != nil {
			log.Printf("토픽 구독 실패: %v", token.Error())
		} else {
			log.Printf("토픽 '%s'을(를) 성공적으로 구독했습니다.", mqttTopic)
		}
	}

	// 연결 실패 핸들러 추가
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("MQTT 연결이 끊어졌습니다: %v", err)
	}

	// 클라이언트 생성 및 연결
	client := mqtt.NewClient(opts)
	log.Println("MQTT 클라이언트 생성 완료, 연결 시도 중...")

	// 연결 시도 타임아웃 설정 (10초)
	go func() {
		token := client.Connect()
		if token.Wait() && token.Error() != nil {
			log.Printf("MQTT 연결 실패: %v", token.Error())
		}
	}()

	select {
	case <-connectedChan:
		log.Println("MQTT 연결 성공")
	case <-time.After(10 * time.Second):
		log.Println("MQTT 연결 시도 타임아웃 (10초 초과)")
		// 필요한 경우, 재시도 로직 또는 애플리케이션 종료를 추가할 수 있습니다.
	}

	// 시그널 처리: Graceful Shutdown 구현
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("수신한 시그널: %v, 애플리케이션 종료 중...", sig)
		client.Disconnect(250) // MQTT 클라이언트 종료
		db.Close()             // 데이터베이스 연결 종료
		os.Exit(0)
	}()

	// 애플리케이션이 종료되지 않도록 대기
	select {}
}

// messageHandler는 MQTT 메시지를 처리하는 핸들러를 반환합니다.
func messageHandler(db *sql.DB) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("수신한 메시지 - 토픽: %s, 페이로드: %s", msg.Topic(), string(msg.Payload()))

		var data IoTData
		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Printf("JSON 파싱 오류: %v, 메시지: %s", err, string(msg.Payload()))
			return
		}

		timestamp, err := time.Parse("2006-01-02 15:04:05", data.Timestamp)
		if err != nil {
			log.Printf("타임스탬프 파싱 오류: %v, 타임스탬프: %s", err, data.Timestamp)
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
	}
}
