CREATE TABLE iot_data (
                          device VARCHAR(50),
                          timestamp TIMESTAMP,
                          pro_ver INT,
                          minor_ver INT,
                          sn BIGINT,
                          model VARCHAR(50),
                          tyield FLOAT,
                          dyield FLOAT,
                          pf FLOAT,
                          pmax INT,
                          pac INT,
                          sac INT,
                          uab INT,
                          ubc INT,
                          uca INT,
                          ia INT,
                          ib INT,
                          ic INT,
                          freq INT,
                          tmod FLOAT,
                          tamb FLOAT,
                          mode VARCHAR(20),
                          qac INT,
                          bus_capacitance FLOAT,
                          ac_capacitance FLOAT,
                          pdc FLOAT,
                          pmax_lim FLOAT,
                          smax_lim FLOAT
);

select * from iot_data;


