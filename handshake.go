package mtproto

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"github.com/xelaj/go-dry"

	ige "github.com/lonesta/mtproto/aes_ige"
	"github.com/lonesta/mtproto/keys"
	"github.com/lonesta/mtproto/serialize"
)

// https://tlgrm.ru/docs/mtproto/auth_key
// https://core.telegram.org/mtproto/auth_key
func (m *MTProto) makeAuthKey() error {
	m.serviceModeActivated = true
	nonceFirst := serialize.RandomInt128()
	res, err := m.ReqPQ(nonceFirst)
	if err != nil {
		return errors.Wrap(err, "requesting first pq")
	}

	if nonceFirst.Cmp(res.Nonce.Int) != 0 {
		return errors.New("handshake: Wrong nonce")
	}
	found := false
	for _, b := range res.Fingerprints {
		if uint64(b) == binary.LittleEndian.Uint64(keys.RSAFingerprint(m.publicKey)) {
			found = true
			break
		}
	}
	if !found {
		return errors.New("handshake: Can't find fingerprint")
	}

	// (encoding) p_q_inner_data
	pq := big.NewInt(0).SetBytes(res.Pq)
	p, q := splitPQ(pq)
	nonceSecond := serialize.RandomInt256()
	nonceServer := res.ServerNonce

	message := (&serialize.PQInnerData{
		Pq:          res.Pq,
		P:           p.Bytes(),
		Q:           q.Bytes(),
		Nonce:       nonceFirst,
		ServerNonce: nonceServer,
		NewNonce:    nonceSecond,
	}).Encode()

	hashAndMsg := make([]byte, 255)
	copy(hashAndMsg, append(dry.Sha1(string(message)), message...))

	encryptedMessage := doRSAencrypt(hashAndMsg, m.publicKey)

	keyFingerprint := int64(binary.LittleEndian.Uint64(keys.RSAFingerprint(m.publicKey)))
	dhResponse, err := m.ReqDHParams(nonceFirst, nonceServer, p.Bytes(), q.Bytes(), keyFingerprint, encryptedMessage)
	if err != nil {
		return errors.Wrap(err, "sending ReqDHParams")
	}
	dhParams, ok := dhResponse.(*serialize.ServerDHParamsOk)
	if !ok {
		return errors.New("handshake: Need ServerDHParamsOk")
	}

	if nonceFirst.Cmp(dhParams.Nonce.Int) != 0 {
		return errors.New("handshake: Wrong nonce")
	}
	if nonceServer.Cmp(dhParams.ServerNonce.Int) != 0 {
		return errors.New("handshake: Wrong server_nonce")
	}

	// проверку по хешу, удаление рандомных байт происходит в этой функции
	decodedMessage := ige.DecryptMessageWithTempKeys(dhParams.EncryptedAnswer, nonceSecond.Int, nonceServer.Int)
	buf := serialize.NewDecoder(decodedMessage)
	data := buf.PopObj()

	dhi, ok := data.(*serialize.ServerDHInnerData)
	if !ok {
		return errors.New("Handshake: Need server_DH_inner_data")
	}
	if nonceFirst.Cmp(dhi.Nonce.Int) != 0 {
		return errors.New("Handshake: Wrong nonce")
	}
	if nonceServer.Cmp(dhi.ServerNonce.Int) != 0 {
		return errors.New("Handshake: Wrong server_nonce")
	}

	// вот это видимо как раз и есть часть диффи хеллмана, поэтому просто оставим как есть надеюсь сработает
	_, g_b, g_ab := makeGAB(dhi.G, big.NewInt(0).SetBytes(dhi.GA), big.NewInt(0).SetBytes(dhi.DhPrime))

	authKey := g_ab.Bytes()
	if authKey[0] == 0 {
		authKey = authKey[1:]
	}

	m.SetAuthKey(authKey)

	// что это я пока не знаю, видимо какой то очень специфичный способ сгенерить ключи
	t4 := make([]byte, 32+1+8)
	copy(t4[0:], nonceSecond.Bytes())
	t4[32] = 1
	copy(t4[33:], dry.Sha1Byte(m.GetAuthKey())[0:8])
	nonceHash1 := dry.Sha1Byte(t4)[4:20]
	salt := make([]byte, serialize.LongLen)
	copy(salt, nonceSecond.Bytes()[:8])
	xor(salt, nonceServer.Bytes()[:8])
	m.serverSalt = int64(binary.LittleEndian.Uint64(salt))

	// (encoding) client_DH_inner_data
	clientDHData := &serialize.ClientDHInnerData{nonceFirst, nonceServer, 0, g_b.Bytes()}

	encryptedMessage = ige.EncryptMessageWithTempKeys(clientDHData.Encode(), nonceSecond.Int, nonceServer.Int)

	dhGenStatus, err := m.SetClientDHParams(nonceFirst, nonceServer, encryptedMessage)
	if err != nil {
		return errors.Wrap(err, "sending clientDHParams")
	}

	dhg, ok := dhGenStatus.(*serialize.DHGenOk)
	if !ok {
		return errors.New("Handshake: Need DHGenOk")
	}
	if nonceFirst.Cmp(dhg.Nonce.Int) != 0 {
		return fmt.Errorf("Handshake: Wrong nonce: %v, %v", nonceFirst, dhg.Nonce)
	}
	if nonceServer.Cmp(dhg.ServerNonce.Int) != 0 {
		return fmt.Errorf("Handshake: Wrong server_nonce: %v, %v", nonceServer, dhg.ServerNonce)
	}
	if !bytes.Equal(nonceHash1, dhg.NewNonceHash1.Bytes()) {
		return fmt.Errorf(
			"handshake: Wrong new_nonce_hash1: %v, %v",
			hex.EncodeToString(nonceHash1),
			hex.EncodeToString(dhg.NewNonceHash1.Bytes()),
		)
	}
	m.serviceModeActivated = false

	// (all ok)
	err = m.SaveSession()
	return errors.Wrap(err, "saving session")
}
