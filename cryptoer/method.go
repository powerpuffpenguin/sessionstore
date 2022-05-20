package cryptoer

import "sync"

var signingMethods = map[string]func() SigningMethod{}
var signingMethodLock = new(sync.RWMutex)

// 簽名和驗證簽名
type SigningMethod interface {
	// 如果簽名有效返回 nil
	Verify(key, value []byte, signature string) error
	// 爲 value 簽名
	Sign(key, value []byte) (string, error)
	// 返回註冊的簽名算法名稱(例如:`HS256`)
	Alg() string
}

// 註冊簽名 alg 的工廠方法，這通常在 init() 函數中完成
func RegisterSigningMethod(alg string, f func() SigningMethod) {
	signingMethodLock.Lock()
	defer signingMethodLock.Unlock()

	signingMethods[alg] = f
}

// 返回註冊名稱爲 "alg" 的簽名工廠方法
func GetSigningMethod(alg string) (method SigningMethod) {
	signingMethodLock.RLock()
	defer signingMethodLock.RUnlock()

	if methodF, ok := signingMethods[alg]; ok {
		method = methodF()
	}
	return
}
