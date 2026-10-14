package agentbridge

import (
	"sync"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

var progressMessageKeys = struct {
	once   sync.Once
	byCode map[ProgressCode]string
}{
	byCode: map[ProgressCode]string{},
}

func progressMessageKey(code ProgressCode) string {
	progressMessageKeys.once.Do(func() {
		catalog, err := progressmessage.Catalog()
		if err != nil {
			return
		}
		for _, message := range catalog.Messages {
			progressMessageKeys.byCode[ProgressCode(message.Code)] = message.Key
		}
	})
	return progressMessageKeys.byCode[code]
}
