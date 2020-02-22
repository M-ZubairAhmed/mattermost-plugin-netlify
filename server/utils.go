package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/mattermost/mattermost-server/v5/model"
)

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func encrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	msg := pad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], msg)
	finalMsg := base64.URLEncoding.EncodeToString(ciphertext)
	return finalMsg, nil
}

func decrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multipe of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := unpad(msg)
	if err != nil {
		return "", err
	}

	return string(unpadMsg), nil
}

// sendBotEphemeralPostWithMessage : Sends an ephemeral bot post to the channel from which slash command was executed
func (p *Plugin) sendBotEphemeralPostWithMessage(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	p.API.SendEphemeralPost(args.UserId, post)
}

func (p *Plugin) setNetlifyUserAccessTokenToStore(accessToken string, userID string) error {
	// Get the encryption from the plugin settings
	encryptionKey := []byte(p.getConfiguration().EncryptionKey)

	// Encrypt the access token with the encryption key
	encryptedAccessToken, err := encrypt(encryptionKey, accessToken)
	if err != nil {
		return err
	}

	// Construct a unique identifier such as "username_token"
	kvStoreKeyIdentifier := userID + NetlifyAuthTokenKVIdentifier

	// Store the encrypted access token in KV store
	err = p.API.KVSet(kvStoreKeyIdentifier, []byte(encryptedAccessToken))
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) sendBotPostOnDM(userID string, message string) *model.AppError {
	// Get the Bot Direct Message channel
	directChannel, err := p.API.GetDirectChannel(userID, p.BotUserID)
	if err != nil {
		return err
	}

	// Construct the Post message
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: directChannel.Id,
		Message:   message,
	}

	// Send the Post
	_, err = p.API.CreatePost(post)
	if err != nil {
		return err
	}

	return nil
}
