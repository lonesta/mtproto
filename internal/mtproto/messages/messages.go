package messages

// messages implements encoding and decoding messages in mtproto
// messages can be encoded end decoded

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
	"github.com/xelaj/go-dry"

	"github.com/lonesta/mtproto/encoding/tl"
	ige "github.com/lonesta/mtproto/internal/aes_ige"
	"github.com/lonesta/mtproto/utils"
)

// CommonMessage это сообщение (зашифрованое либо открытое) которыми общаются между собой клиент и сервер
type CommonMessage interface {
	GetMsg() []byte
	GetMsgID() int
	GetSeqNo() int
}

type EncryptedMessage struct {
	Msg         []byte
	MsgID       int64
	AuthKeyHash []byte

	Salt      int64
	SessionID int64
	SeqNo     int32
	MsgKey    []byte
}

func (msg *EncryptedMessage) Serialize(client MessageInformator, requireToAck bool) ([]byte, error) {
	obj := serializePacket(client, msg.Msg, msg.MsgID, requireToAck)
	encryptedData, err := ige.Encrypt(obj, client.GetAuthKey())
	if err != nil {
		return nil, errors.Wrap(err, "encrypting")
	}

	buf := bytes.NewBuffer(nil)

	e := tl.NewEncoder(buf)
	e.PutRawBytes(utils.AuthKeyHash(client.GetAuthKey()))
	e.PutRawBytes(ige.MessageKey(obj))
	e.PutRawBytes(encryptedData)

	return buf.Bytes(), nil
}

func DeserializeEncryptedMessage(data, authKey []byte) (*EncryptedMessage, error) {
	msg := new(EncryptedMessage)

	buf := bytes.NewBuffer(data)
	d := tl.NewDecoder(buf)
	keyHash := d.PopRawBytes(tl.LongLen)
	if !bytes.Equal(keyHash, utils.AuthKeyHash(authKey)) {
		return nil, errors.New("wrong encryption key")
	}
	msg.MsgKey = d.PopRawBytes(tl.Int128Len) // msgKey это хэш от расшифрованного набора байт, последние 16 символов
	encryptedData := d.PopRawBytes(len(data) - (tl.LongLen + tl.Int128Len))

	decrypted, err := ige.Decrypt(encryptedData, authKey, msg.MsgKey)
	if err != nil {
		return nil, errors.Wrap(err, "decrypting message")
	}
	buf = bytes.NewBuffer(decrypted)
	d = tl.NewDecoder(buf)
	msg.Salt = d.PopLong()
	msg.SessionID = d.PopLong()
	msg.MsgID = d.PopLong()
	msg.SeqNo = d.PopInt()
	messageLen := d.PopInt()

	if len(decrypted) < int(messageLen)-(tl.LongLen+tl.LongLen+tl.LongLen+tl.WordLen+tl.WordLen) {
		return nil, fmt.Errorf("message is smaller than it's defining: have %v, but messageLen is %v", len(decrypted), messageLen)
	}

	mod := msg.MsgID & 3
	if mod != 1 && mod != 3 {
		return nil, fmt.Errorf("wrong bits of message_id: %d", mod)
	}

	// этот кусок проверяет валидность данных по ключу
	trimed := decrypted[0 : 32+messageLen] // суммарное сообщение, после расшифровки
	if !bytes.Equal(dry.Sha1Byte(trimed)[4:20], msg.MsgKey) {
		return nil, errors.New("wrong message key, can't trust to sender")
	}
	msg.Msg = d.PopRawBytes(int(messageLen))

	return msg, nil
}

func (msg *EncryptedMessage) GetMsg() []byte {
	return msg.Msg
}

func (msg *EncryptedMessage) GetMsgID() int {
	return int(msg.MsgID)
}

func (msg *EncryptedMessage) GetSeqNo() int {
	return int(msg.SeqNo)
}

type UnencryptedMessage struct {
	Msg   []byte
	MsgID int64
}

func (msg *UnencryptedMessage) Serialize(client MessageInformator) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	e := tl.NewEncoder(buf)
	// authKeyHash, always 0 if unencrypted
	e.PutLong(0)
	e.PutLong(msg.MsgID)
	e.PutInt(int32(len(msg.Msg)))
	e.PutRawBytes(msg.Msg)
	return buf.Bytes(), nil
}

func DeserializeUnencryptedMessage(data []byte) (*UnencryptedMessage, error) {
	msg := new(UnencryptedMessage)
	d := tl.NewDecoder(bytes.NewBuffer(data))
	_ = d.PopRawBytes(tl.LongLen) // authKeyHash, always 0 if unencrypted

	msg.MsgID = d.PopLong()

	mod := msg.MsgID & 3
	if mod != 1 && mod != 3 {
		return nil, fmt.Errorf("Wrong bits of message_id: %#v", uint64(mod))
	}

	messageLen := d.PopUint()
	if len(data)-(tl.LongLen+tl.LongLen+tl.WordLen) != int(messageLen) {
		fmt.Println(len(data), int(messageLen), int(messageLen+(tl.LongLen+tl.LongLen+tl.WordLen)))
		return nil, fmt.Errorf("message not equal defined size: have %v, want %v", len(data), messageLen)
	}

	var err error
	msg.Msg, err = d.GetRestOfMessage()
	if err != nil {
		return nil, errors.Wrap(err, "getting real message")
	}

	return msg, nil
}

func (msg *UnencryptedMessage) GetMsg() []byte {
	return msg.Msg
}

func (msg *UnencryptedMessage) GetMsgID() int {
	return int(msg.MsgID)
}

func (msg *UnencryptedMessage) GetSeqNo() int {
	return 0
}

//------------------------------------------------------------------------------------------

// MessageInformator нужен что бы отдавать информацию о текущей сессии для сериализации сообщения
// по факту это *MTProto структура
type MessageInformator interface {
	GetSessionID() int64
	GetLastSeqNo() int32
	GetServerSalt() int64
	GetAuthKey() []byte
}

func serializePacket(client MessageInformator, msg []byte, messageID int64, requireToAck bool) []byte {
	buf := bytes.NewBuffer(nil)
	d := tl.NewEncoder(buf)

	saltBytes := make([]byte, tl.LongLen)
	binary.LittleEndian.PutUint64(saltBytes, uint64(client.GetServerSalt()))
	d.PutRawBytes(saltBytes)
	d.PutLong(client.GetSessionID())
	d.PutLong(messageID)
	if requireToAck { // не спрашивай, как это работает
		d.PutInt(client.GetLastSeqNo() | 1) // почему тут добавляется бит не ебу
	} else {
		d.PutInt(client.GetLastSeqNo())
	}
	d.PutInt(int32(len(msg)))
	d.PutRawBytes(msg)
	return buf.Bytes()
}
