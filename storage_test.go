package osinredis

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/RangelReale/osin"
	"github.com/garyburd/redigo/redis"
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
	assert.True(t, reflect.DeepEqual(loadData, authorizeData))
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

	loadData, err := storage.LoadAccess(accessData.AccessToken)
	assert.Equal(t, loadData, accessData)
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
	assert.Equal(t, loadData, accessData)
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
