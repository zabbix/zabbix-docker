CREATE TABLE zabbix.history_uint
(
  itemid UInt64,
  clock_ns DateTime64(9),
  value UInt64
)
ENGINE = MergeTree()
PARTITION BY toDate(clock_ns)
PRIMARY KEY (itemid, clock_ns)
TTL clock_ns + toIntervalSecond(2678400);
