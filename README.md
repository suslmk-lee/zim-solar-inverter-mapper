# zim-solar-inverter-mapper
삼환에서 제공하는 태양광데이터(IoT)의 Inverter에서 보내주는 데이터를 받아 처리하는 KubeEdge Mapper 컨테이너

```shell
## mqtt broker 실행
mosquitto_sub -h localhost -p 1883 -t iot/data
```

```shell
## test로 publish 수행
mosquitto_pub -h localhost -p 1883 -t iot/data -m '{
  "Device": "IoT_002",
  "Timestamp": "2024-11-12 15:18:10",
  "ProVer": 12345,
  "MinorVer": 12345,
  "SN": 123456789012345,
  "model": "mo-2093",
  "Status": {
    "Tyield": 4929.23,
    "Dyield": 432.77,
    "PF": 23.81,
    "Pmax": 12345,
    "Pac": 12345,
    "Sac": 12345,
    "Uab": 12345,
    "Ubc": 12345,
    "Uca": 12345,
    "Ia": 12345,
    "Ib": 12345,
    "Ic": 12345,
    "Freq": 12345,
    "Tmod": 32.8,
    "Tamb": 27.4,
    "Mode": "Running",
    "Qac": 1189,
    "BusCapacitance": 16.2,
    "AcCapacitance": 17.1,
    "Pdc": 42.12,
    "PmaxLim": 46.82,
    "SmaxLim": 39.92
  }
}'                                                                                              <....
```