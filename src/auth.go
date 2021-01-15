package akyuu

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
)

// god i need a database

const TokenHeaderKey = "X-AUTH-TOKEN"
const TokenMod = 1 << 32

var TokenOffset = rand.Int()

type TokenId string

type Token struct {
	Activated bool    `json:"activated"`
	ID        TokenId `json:"id"`
	Name      string  `json:"name"`
	Issuer    TokenId `json:"issuer"`
}

func GenerateTokenId() TokenId {
	val := int(math.Abs(float64((rand.Int() + TokenOffset) % TokenMod)))
	return TokenId(fmt.Sprint(val))
}

type TokenReg struct {
	StoragePath  string    `json:"config-path"`
	ValidIssuers []TokenId `json:"issuers"`
	Tokens       []Token   `json:"tokens"`
}

func NewTokenReg(storagePath string) (TokenReg, error) {
	reg := TokenReg{
		StoragePath: storagePath,
	}
	if err := reg.Load(); err != nil {
		return TokenReg{}, err
	}
	return reg, nil
}

func (r *TokenReg) Load() error {
	file, err := os.Open(r.StoragePath)
	if err != nil {
		return err
	}
	defer file.Close()
	d, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(d, &r)
	if err != nil {
		return err
	}
	return nil
}

func (r TokenReg) Write() error {
	m, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(r.StoragePath, m, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (r TokenReg) GetTokenById(id TokenId) (Token, bool) {
	for _, token := range r.Tokens {
		if token.ID == id {
			return token, true
		}
	}
	return Token{}, false
}

func (r TokenReg) VerifyIssuerPerms(token Token) bool {
	for _, issuer := range r.ValidIssuers {
		if token.ID == issuer {
			return true
		}
	}
	return false
}

func (r TokenReg) VerifyValidIssuer(token Token) bool {
	for _, i := range r.ValidIssuers {
		if i == token.Issuer {
			return true
		}
	}
	return false
}

func (r TokenReg) VerifyToken(id TokenId) (Token, bool) {
	token, present := r.GetTokenById(id)
	if !present || !token.Activated {
		return Token{}, false
	}
	return token, true
}

func (r *TokenReg) CreateToken(tokenName string, issuerToken Token) (Token, error) {
	newToken := Token{
		Activated: true,
		ID:        GenerateTokenId(),
		Name:      tokenName,
		Issuer:    issuerToken.ID,
	}
	r.Tokens = append(r.Tokens, newToken)
	if err := r.Write(); err != nil {
		r.Tokens = r.Tokens[:len(r.Tokens)-1]
		return Token{}, err
	}
	return newToken, nil
}

func (r *TokenReg) ListTokens(onlyActivated bool) []Token {
	tokens := make([]Token, 0)
	for _, token := range r.Tokens {
		if onlyActivated && !token.Activated {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func (r *TokenReg) UpdateToken(id TokenId, newToken Token) error {
	oldToken, present := r.GetTokenById(id)
	if !present {
		return fmt.Errorf("Token %s not found", id)
	}
	for i, token := range r.Tokens {
		if token.ID == id {
			r.Tokens[i] = newToken
			if err := r.Write(); err != nil {
				r.Tokens[i] = oldToken
				return err
			}
			break
		}
	}
	return nil
}
