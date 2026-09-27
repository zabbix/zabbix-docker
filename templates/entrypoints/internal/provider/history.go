package provider

// HistoryPrefix is the variable prefix of one history storage provider.
const HistoryPrefix = "ZBX_HISTORYPROVIDER"

// clickHouseHistory describes the options handled by the image and frontend.
var clickHouseHistory = schema{
	valueTypes: true,
	files:      tlsFiles,
	web:        []string{"provider", "types", "url", "db", "username", "password"},
	secrets:    true,
}

// History describes the server HistoryProvider parameter and the frontend
// history providers. The frontend also accepts one unindexed JSON array.
var History = &parameter{
	Prefix:        HistoryPrefix,
	Plural:        "ZBX_HISTORYPROVIDERS",
	pluralInput:   true,
	providerError: "provider must be clickhouse or elasticsearch",
	reserved:      historyReserved,
	providers: map[string]schema{
		"clickhouse": clickHouseHistory,
		"elasticsearch": {
			valueTypes: true,
			web:        []string{"provider", "types", "url"},
		},
	},
}

// historyReserved adds the JSON name of the native value_types option to the
// directories the image owns. Options such as source_ip or log_slow_queries are
// valid per provider and are left to the server.
var historyReserved = func() map[string]string {
	reserved := make(map[string]string, len(managedLocations)+1)
	for key, reason := range managedLocations {
		reserved[key] = reason
	}
	reserved["value_types"] = "use types instead of value_types"
	return reserved
}()
