// Package auth는 비밀번호 해시와 세션 토큰 같은 보안 기본 요소를 모은다.
// HTTP나 DB를 모르는 순수 함수들이라 테스트와 재사용이 쉽다.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Cost는 bcrypt 작업 계수다. 1 올릴 때마다 계산 시간이 2배가 된다.
// 테스트에서는 속도를 위해 bcrypt.MinCost로 낮춘다.
var Cost = bcrypt.DefaultCost

// HashPassword는 솔트가 포함된 bcrypt 해시를 만든다. 같은 비밀번호라도 매번 결과가 다르다.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), Cost)
	return string(h), err
}

// dummyHash는 존재하지 않는 사용자로 로그인할 때 비교 대상으로 쓴다.
// 사용자가 없다고 바로 반환하면 응답이 빨라져서, 응답 시간으로 가입된 이메일을 알아낼 수 있다(user enumeration).
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

// CheckPassword는 해시와 비밀번호가 일치하는지 확인한다.
// hash가 빈 문자열이면(사용자 없음) 더미 해시와 비교해 같은 시간을 소모한 뒤 false를 반환한다.
func CheckPassword(hash, password string) bool {
	if hash == "" {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ErrPasswordTooLong: bcrypt는 72바이트까지만 사용하므로 그 이상은 거부한다.
var ErrPasswordTooLong = errors.New("password must be at most 72 bytes")

// ValidatePassword는 최소한의 비밀번호 규칙을 검사한다.
func ValidatePassword(p string) error {
	if len(p) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(p) > 72 {
		return ErrPasswordTooLong
	}
	return nil
}

// NewToken은 32바이트(256비트) 암호학적 난수를 URL-safe 문자열로 만든다.
// math/rand가 아니라 crypto/rand를 써야 예측할 수 없다.
func NewToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // Go 1.24+: crypto/rand.Read는 실패하지 않는다
	return base64.RawURLEncoding.EncodeToString(b)
}

// HashToken은 DB 저장용 토큰 해시를 만든다.
// 토큰 자체가 고엔트로피 난수라 bcrypt 같은 느린 해시가 필요 없다.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
