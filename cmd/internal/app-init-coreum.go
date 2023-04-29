package internal

import (
	"crypto_processor/modules/service"
	"database/sql"
)

// ToDo inti in the system
func InitProcessorCoreum(blockchain string, db *sql.DB) service.CryptoProcessor {
	return service.CryptoProcessor()
}
