CREATE TABLE zabbix.history_json
(
  itemid UInt64,
  clock_ns DateTime64(9),
  value JSON,
  value_str String
)
ENGINE = MergeTree()
PARTITION BY toDate(clock_ns)
PRIMARY KEY (itemid, clock_ns)
TTL clock_ns + toIntervalSecond(2678400)
