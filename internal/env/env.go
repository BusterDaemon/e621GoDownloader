package env

import (
	"os"
	"strconv"
)

func GetEnvData(waitTime *int, maxPostPages *uint,
	outDir *string, proxyUrl *string, dbPath *string) error {
	var (
		proxy    = os.Getenv("E621_PROXY_URL")
		out      = os.Getenv("E621_OUTPUT_DIRECTORY")
		maxPages = os.Getenv("E621_MAX_PAGES")
		idleTime = os.Getenv("E621_WAIT_TIME")
		db       = os.Getenv("E621_DB_PATH")
	)

	if proxy != "" {
		*proxyUrl = proxy
	}
	if out != "" {
		*outDir = out
	}

	if maxPages != "" {
		maxPagesI, err := strconv.ParseUint(maxPages, 10, 64)
		if err != nil {
			return err
		}
		*maxPostPages = uint(maxPagesI)
	}

	if idleTime != "" {
		idleTimeI, err := strconv.Atoi(idleTime)
		if err != nil {
			return err
		}
		*waitTime = idleTimeI
	}

	if db != "" {
		*dbPath = db
	}

	return nil
}
