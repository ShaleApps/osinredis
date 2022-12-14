package osinredis

import (
	"os"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/openshift/osin"
	"github.com/stretchr/testify/assert"
)

var (
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			redisAddr := os.Getenv("REDIS_ADDR")
			if redisAddr == "" {
				redisAddr = ":6379"
			}
			conn, err := redis.Dial("tcp", redisAddr)
			if err != nil {
				panic(err)
			}
			return conn, nil
		},
	}
)

func init() {
	flushAll()
}

func initTestStorage() *Storage {
	return New(pool, "test123")
}

func flushAll() {
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHALL")
	if err != nil {
		panic(err)
	}
}

func newClient() *osin.DefaultClient {
	return &osin.DefaultClient{Id: "clientID", Secret: "secret", RedirectUri: "http://localhost/", UserData: make(map[string]interface{})}
}

func newAuthorizeData(client *osin.DefaultClient) *osin.AuthorizeData {
	return &osin.AuthorizeData{
		Client:      client,
		Code:        "8888",
		ExpiresIn:   3600,
		CreatedAt:   time.Now(),
		RedirectUri: "http://localhost/",
	}
}

func newAccessData(authorizeData *osin.AuthorizeData) *osin.AccessData {
	return &osin.AccessData{
		Client:        authorizeData.Client,
		AuthorizeData: authorizeData,
		AccessToken:   "8888",
		RefreshToken:  "r8888",
		ExpiresIn:     3600,
		CreatedAt:     time.Now(),
	}
}

func isEqualAuthorizeData(s1, s2 interface{}) bool {
	if _, ok := s1.(*osin.AuthorizeData); ok {
		if _, ok := s2.(*osin.AuthorizeData); ok {
			if s1.(*osin.AuthorizeData).RedirectUri == s2.(*osin.AuthorizeData).RedirectUri {
				if s1.(*osin.AuthorizeData).ExpiresIn == s2.(*osin.AuthorizeData).ExpiresIn {
					return s1.(*osin.AuthorizeData).Code == s2.(*osin.AuthorizeData).Code
				} else {
					return false
				}
			}
			return false
		}
		return false
	}
	return false
}

func isEqualAccessData(s1, s2 interface{}) bool {
	if _, ok := s1.(*osin.AccessData); ok {
		if _, ok := s2.(*osin.AccessData); ok {
			if s1.(*osin.AccessData).AccessToken == s2.(*osin.AccessData).AccessToken {
				if s1.(*osin.AccessData).RefreshToken == s2.(*osin.AccessData).RefreshToken {
					return s1.(*osin.AccessData).ExpiresIn == s2.(*osin.AccessData).ExpiresIn
				} else {
					return false
				}
			}
			return false
		}
		return false
	}
	return false
}

func TestCreateClient(t *testing.T) {
	flushAll()

	storage := initTestStorage()
	client := newClient()
	assert.NoError(t, storage.CreateClient(client))
}

func TestGetClient(t *testing.T) {
	flushAll()

	storage := initTestStorage()
	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	clientFound, err := storage.GetClient(client.GetId())
	assert.NoError(t, err)
	assert.Equal(t, client, clientFound)
}

func TestGetClientNotFound(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	clientFound, err := storage.GetClient("notthere")
	assert.NoError(t, err)
	assert.Nil(t, clientFound)
}

func TestUpdateClient(t *testing.T) {
	flushAll()

	storage := initTestStorage()
	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	client.Secret = "secret_changed"
	client.RedirectUri = "http://localhost/changed"

	assert.NoError(t, storage.UpdateClient(client))

	clientFound, err := storage.GetClient(client.GetId())
	assert.NoError(t, err)
	assert.Equal(t, clientFound, client)
}

func TestDeleteClient(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	err := storage.DeleteClient(client)
	assert.NoError(t, err)
}

func TestSaveAuthorize(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	authorizeData := newAuthorizeData(client)
	assert.NoError(t, storage.SaveAuthorize(authorizeData))
}

func TestLoadAuthorizeNonExistent(t *testing.T) {
	flushAll()

	storage := initTestStorage()
	loadData, err := storage.LoadAuthorize("nonExistentCode")
	assert.Nil(t, loadData)
	assert.NoError(t, err)
}

func TestLoadAuthorize(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	authorizeData := newAuthorizeData(client)
	assert.NoError(t, storage.SaveAuthorize(authorizeData))

	loadData, err := storage.LoadAuthorize(authorizeData.Code)
	assert.NoError(t, err)
	assert.True(t, isEqualAuthorizeData(loadData, authorizeData))
}

func TestRemoveAuthorizeNonExistent(t *testing.T) {
	flushAll()

	storage := initTestStorage()
	assert.NoError(t, storage.RemoveAuthorize("nonExistentCode"))
}

func TestRemoveAuthorize(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	authorizeData := newAuthorizeData(client)
	assert.NoError(t, storage.SaveAuthorize(authorizeData))
	assert.NoError(t, storage.RemoveAuthorize(authorizeData.Code))

	loadData, err := storage.LoadAuthorize(authorizeData.Code)
	assert.Nil(t, loadData)
	assert.NoError(t, err)
}

func TestSaveAccess(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	authorizeData := newAuthorizeData(client)
	assert.NoError(t, storage.SaveAuthorize(authorizeData))

	accessData := newAccessData(authorizeData)
	assert.NoError(t, storage.SaveAccess(accessData))
}

func TestLoadAccessNonExistent(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	loadData, err := storage.LoadAccess("nonExistentToken")
	assert.Nil(t, loadData)
	assert.NoError(t, err)
}

func TestLoadAccess(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	authorizeData := newAuthorizeData(client)
	assert.NoError(t, storage.SaveAuthorize(authorizeData))

	accessData := newAccessData(authorizeData)
	assert.NoError(t, storage.SaveAccess(accessData))

	time.Sleep(1 * time.Second)

	loadData, err := storage.LoadAccess(accessData.AccessToken)
	assert.NotEqual(t, loadData.ExpiresIn, accessData.ExpiresIn)

	loadData.ExpiresIn = accessData.ExpiresIn
	assert.True(t, isEqualAccessData(loadData, accessData))
	assert.NoError(t, err)
}

func TestRemoveAccessNonExistent(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	err := storage.RemoveAccess("nonExistentToken")
	assert.Error(t, err)
}

func TestRemoveAccess(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	authorizeData := newAuthorizeData(client)
	assert.NoError(t, storage.SaveAuthorize(authorizeData))

	accessData := newAccessData(authorizeData)
	assert.NoError(t, storage.SaveAccess(accessData))
	assert.NoError(t, storage.RemoveAccess(accessData.AccessToken))

	loadData, err := storage.LoadAccess(accessData.AccessToken)
	assert.Nil(t, loadData)
	assert.NoError(t, err)
}

func TestLoadRefreshNonExistent(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	loadData, err := storage.LoadRefresh("nonExistentToken")
	assert.Nil(t, loadData)
	assert.NoError(t, err)
}

func TestLoadRefresh(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	authorizeData := newAuthorizeData(client)
	assert.NoError(t, storage.SaveAuthorize(authorizeData))

	accessData := newAccessData(authorizeData)
	assert.NoError(t, storage.SaveAccess(accessData))

	loadData, err := storage.LoadRefresh(accessData.RefreshToken)
	assert.NoError(t, err)
	assert.True(t, isEqualAccessData(loadData, accessData))
}

func TestRemoveRefreshNonExistent(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	err := storage.RemoveRefresh("nonExistentToken")
	assert.Error(t, err)
}

func TestRemoveRefresh(t *testing.T) {
	flushAll()

	storage := initTestStorage()

	client := newClient()
	assert.NoError(t, storage.CreateClient(client))

	authorizeData := newAuthorizeData(client)
	assert.NoError(t, storage.SaveAuthorize(authorizeData))

	accessData := newAccessData(authorizeData)
	assert.NoError(t, storage.SaveAccess(accessData))

	err := storage.RemoveRefresh(accessData.RefreshToken)
	assert.NoError(t, err)

	loadData, err := storage.LoadRefresh(accessData.RefreshToken)
	assert.Nil(t, loadData)
	assert.NoError(t, err)
}
