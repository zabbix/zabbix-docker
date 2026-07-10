CREATE TABLE zabbix.history_log
(
  itemid UInt64,
  clock_ns DateTime64(9),
  value String,
  source String,
  severity Int32,
  logeventid Int32,
  timestamp Int64
)
ENGINE = MergeTree()
PARTITION BY toDate(clock_ns)
PRIMARY KEY (itemid, clock_ns)
TTL clock_ns + toIntervalSecond(2678400);
