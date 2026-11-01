package storageprovider

const historyPrefix = "ZBX_HISTORYPROVIDER"

var clickHouseHistory = providerDefinition{
	valueTypes: true,
	files:      tlsFiles,
	web:        []string{"provider", "types", "url", "db", "username", "password"},
	secrets:    true,
}

// History describes the server HistoryProvider parameter and the frontend
// history providers. The frontend also accepts one unindexed JSON array.
var History = &Parameter{
	Prefix:        historyPrefix,
	Plural:        "ZBX_HISTORYPROVIDERS",
	pluralInput:   true,
	providerError: "provider must be clickhouse, elasticsearch or sql",
	reserved:      map[string]string{"value_types": "use types instead of value_types"},
	providers: map[string]providerDefinition{
		"clickhouse": clickHouseHistory,
		"elasticsearch": {
			valueTypes: true,
			web:        []string{"provider", "types", "url"},
		},
		"sql": {valueTypes: true},
	},
}
