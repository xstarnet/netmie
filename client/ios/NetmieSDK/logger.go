//go:build ios

package NetmieSDK

import (
	"github.com/netbirdio/netbird/util"
)

// InitializeLog initializes the log file.
func InitializeLog(logLevel string, filePath string) error {
	return util.InitLog(logLevel, filePath)
}
