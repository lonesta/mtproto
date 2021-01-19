package mtproto

import (
	"github.com/lonesta/mtproto/serialize"
)

// это неофициальная информация, но есть подозрение, что список датацентров АБСОЛЮТНО идентичный для всех
// приложений. Несмотря на это, любой клиент ОБЯЗАН явно указывать список датацентров, ради надежности.
// данный список лишь эксперементальный и не является частью протокола.
var defaultDCList = map[int]string{
	1: "149.154.175.58:443",
	2: "149.154.167.50:443",
	3: "149.154.175.100:443",
	4: "149.154.167.91:443",
	5: "91.108.56.151:443",
}

func MessageRequireToAck(msg serialize.TL) bool {
	switch msg.(type) {
	case /**serialize.Ping,*/ *serialize.MsgsAck:
		return false
	default:
		return true
	}
}
