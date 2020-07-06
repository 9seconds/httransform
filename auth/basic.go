package auth

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"github.com/PumpkinSeed/errors"
	"zvelo.io/ttlru"

	"github.com/9seconds/httransform/v2/layers"
)

type basicAuthResult struct {
	reply interface{}
	err   error
}

type basicAuthUserInfo struct {
	user     []byte
	password []byte
}

func (u *basicAuthUserInfo) OK(user, password []byte) bool {
	userNum := subtle.ConstantTimeCompare(u.user, user)
	passNum := subtle.ConstantTimeCompare(u.password, password)

	return userNum+passNum == 2
}

type basicAuth struct {
	cache ttlru.Cache
	infos []basicAuthUserInfo
}

func (b *basicAuth) Auth(ctx *layers.LayerContext) (interface{}, error) {
	header := ctx.RequestHeaders.Get("proxy-authorization")

	if header == nil {
		return nil, ErrNoAuth
	}

	if item, ok := b.cache.Get(header.Value); ok {
		reply := item.(*basicAuthResult)
		return reply.reply, reply.err
	}

	resp := b.doAuth(header.Value)
	b.cache.Set(header.Value, &resp)

	return resp.reply, resp.err
}

func (b *basicAuth) doAuth(text string) basicAuthResult {
	pos := strings.IndexByte(text, ' ')
	if pos < 0 {
		return basicAuthResult{
			err: ErrBasicAuthMalformed,
		}
	}

	if !strings.EqualFold(text[:pos], "Basic") {
		return basicAuthResult{
			err: ErrBasicAuthScheme,
		}
	}

	for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t') {
		pos++
	}

	decoded, err := base64.StdEncoding.DecodeString(text[pos:])
	if err != nil {
		return basicAuthResult{
			err: errors.Wrap(err, ErrBasicAuthPayload),
		}
	}

	pos = bytes.IndexByte(decoded, ':')
	if pos < 0 {
		return basicAuthResult{
			err: ErrBasicAuthDelimiter,
		}
	}

	found := false
	for idx := range b.infos {
		found = b.infos[idx].OK(decoded[:pos], decoded[pos+1:]) || found
	}

	if found {
		return basicAuthResult{
			reply: string(decoded[:pos]),
		}
	}

	return basicAuthResult{
		err: ErrBasicAuthNoUser,
	}
}

func NewBasicAuth(userPasswords map[string]string) Auth {
	userInfos := make([]basicAuthUserInfo, len(userPasswords))
	idx := 0

	for k, v := range userPasswords {
		userInfos[idx].user = []byte(k)
		userInfos[idx].password = []byte(v)
		idx++
	}

	return &basicAuth{
		cache: ttlru.New(AuthCacheSizeMultiplier*len(userPasswords),
			ttlru.WithTTL(AuthCacheFor)),
		infos: userInfos,
	}
}
