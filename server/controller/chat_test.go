package controller

import (
	"net/http"
	"testing"
	"time"

	"github.com/blendlabs/chatbus/server/model"
	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
	web "github.com/wcharczuk/go-web"
)

type serviceResponseOfUser struct {
	Meta     map[string]interface{} `json:"meta"`
	Response model.User             `json:"response"`
}

type serviceResponseOfSession struct {
	Meta     map[string]interface{} `json:"meta"`
	Response model.Session          `json:"response"`
}

type serviceResponseOfSessions struct {
	Meta     map[string]interface{} `json:"meta"`
	Response []model.Session        `json:"response"`
}

type serviceResponseOfMessage struct {
	Meta     map[string]interface{} `json:"meta"`
	Response model.Message          `json:"response"`
}

type serviceResponseOfMessages struct {
	Meta     map[string]interface{} `json:"meta"`
	Response []model.Message        `json:"response"`
}

func TestChatRegister(t *testing.T) {
	assert := assert.New(t)

	app := web.New()
	app.Register(new(Chat))
	meta, err := app.Mock().WithPathf("/api/sessions").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
}

func TestChatRestore(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	u2 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User2"}
	u3 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User3"}
	assert.Nil(model.DB().CreateInTransaction(u1, tx))
	assert.Nil(model.DB().CreateInTransaction(u2, tx))
	assert.Nil(model.DB().CreateInTransaction(u3, tx))

	s1 := &model.Session{
		UUID:          util.UUIDv4().ToShortString(),
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        u1.ID,
		User:          u1,
	}
	s2 := &model.Session{
		UUID:          util.UUIDv4().ToShortString(),
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        u2.ID,
		User:          u2,
	}
	s3 := &model.Session{
		UUID:          util.UUIDv4().ToShortString(),
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        u3.ID,
		User:          u3,
	}
	assert.Nil(model.DB().CreateInTransaction(s1, tx))
	assert.Nil(model.DB().CreateInTransaction(s2, tx))
	assert.Nil(model.DB().CreateInTransaction(s3, tx))

	c1 := &model.Contacts{Sender: u1.ID, Receiver: u2.ID}
	c2 := &model.Contacts{Sender: u2.ID, Receiver: u1.ID}
	c3 := &model.Contacts{Sender: u1.ID, Receiver: u3.ID}
	c4 := &model.Contacts{Sender: u3.ID, Receiver: u1.ID}
	assert.Nil(model.DB().CreateInTransaction(c1, tx))
	assert.Nil(model.DB().CreateInTransaction(c2, tx))
	assert.Nil(model.DB().CreateInTransaction(c3, tx))
	assert.Nil(model.DB().CreateInTransaction(c4, tx))

	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u3.ID, Body: "Test"}, tx))
	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u3.ID, Body: "Test"}, tx))
	assert.Nil(model.DB().CreateInTransaction(&model.Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u3.ID, Body: "Test"}, tx))

	chat := new(Chat)
	assert.Nil(chat.Restore(tx))
	assert.NotEmpty(chat.Users)
	assert.NotEmpty(chat.Sessions)
	assert.NotEmpty(chat.SessionsByUser)
	assert.NotEmpty(chat.Contacts)
	assert.NotEmpty(chat.MessageQueues)
}

func TestChatCacheUser(t *testing.T) {
	assert := assert.New(t)

	chat := new(Chat)
	chat.cacheUser(&model.User{ID: 99, UUID: "test_user"})
	assert.True(chat.hasCachedUser(99))
}

func TestChatCacheContact(t *testing.T) {
	assert := assert.New(t)

	chat := new(Chat)
	chat.cacheContact(1, 2)
	assert.NotEmpty(chat.Contacts)
	assert.NotZero(chat.Contacts[1].Len())
	assert.True(chat.Contacts[1].Contains(2))
	assert.NotZero(chat.Contacts[2].Len())
	assert.True(chat.Contacts[2].Contains(1))
}

func TestChatCacheSession(t *testing.T) {
	assert := assert.New(t)

	chat := new(Chat)
	chat.cacheSession(&model.Session{
		UUID:          "test_session",
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        1,
		User:          &model.User{ID: 1, UUID: "test_user"},
	})
	assert.NotNil(chat.getCachedSession("test_session"))
	assert.Nil(chat.getCachedSession("not_test_session"))
}

func TestChatCacheSessionByUser(t *testing.T) {
	assert := assert.New(t)
	chat := new(Chat)
	chat.cacheSessionByUser(&model.Session{
		UUID:          "test_session",
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        1,
		User:          &model.User{ID: 1, UUID: "test_user"},
	})
	assert.True(chat.userHasSession(1))
	assert.False(chat.userHasSession(2))
}

func TestChatAddMessageQueue(t *testing.T) {
	assert := assert.New(t)
	chat := new(Chat)
	chat.addMessageQueue(&model.Session{
		UUID:          "test_session",
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        1,
		User:          &model.User{ID: 1, UUID: "test_user"},
	})
	assert.NotNil(chat.MessageQueues)
	assert.NotEmpty(chat.MessageQueues)
	assert.NotNil(chat.MessageQueues[1])
	assert.Nil(chat.MessageQueues[2])
}

func TestChatSetCachedSessionLastActive(t *testing.T) {
	assert := assert.New(t)
	chat := new(Chat)
	chat.cacheSession(&model.Session{
		UUID:       "test_session",
		CreatedUTC: time.Now().UTC(),
		UserID:     1,
		User:       &model.User{ID: 1, UUID: "test_user"},
	})
	assert.True(chat.Sessions["test_session"].LastActiveUTC.IsZero())
	chat.setCachedSessionLastActive("test_session")
	assert.False(chat.Sessions["test_session"].LastActiveUTC.IsZero())
}

func TestChatQueueMessage(t *testing.T) {
	assert := assert.New(t)
	chat := new(Chat)

	session1 := &model.Session{
		UUID:       "test_session",
		CreatedUTC: time.Now().UTC(),
		UserID:     1,
		User:       &model.User{ID: 1, UUID: "test_user1"},
	}
	session2 := &model.Session{
		UUID:       "test_session2",
		CreatedUTC: time.Now().UTC(),
		UserID:     2,
		User:       &model.User{ID: 2, UUID: "test_user2"},
	}

	chat.cacheSession(session1)
	chat.addMessageQueue(session1)
	chat.cacheSession(session2)
	chat.addMessageQueue(session2)

	chat.queueMessage(&model.Message{
		SenderID:   1,
		ReceiverID: 2,
		Body:       "This is a test message.",
	})
	assert.NotEmpty(chat.MessageQueues)
	assert.Len(chat.MessageQueues, 2)
	assert.Equal(1, chat.MessageQueues[1].Len())
	assert.NotZero(1, chat.MessageQueues[2].Len())
}

func TestChatRemoveCachedUser(t *testing.T) {
	assert := assert.New(t)

	chat := new(Chat)
	chat.cacheUser(&model.User{ID: 99, UUID: "test_user"})
	assert.True(chat.hasCachedUser(99))

	chat.removeCachedUser(99)
}

func TestChatRemoveCachedSession(t *testing.T) {
	assert := assert.New(t)

	chat := new(Chat)
	chat.cacheSession(&model.Session{
		UUID:          "test_session",
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        1,
		User:          &model.User{ID: 1, UUID: "test_user"},
	})
	assert.NotNil(chat.getCachedSession("test_session"))
	chat.removeCachedSession("test_session")
	assert.Nil(chat.getCachedSession("test_session"))
}

func TestChatRemoveCachedSessionByUser(t *testing.T) {
	assert := assert.New(t)

	chat := new(Chat)
	s1 := &model.Session{
		UUID:          "test_session",
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        1,
		User:          &model.User{ID: 1, UUID: "test_user"},
	}
	chat.cacheSessionByUser(s1)
	assert.True(chat.userHasSession(1))
	chat.removeCachedSessionByUser(s1)
	assert.False(chat.userHasSession(1))
}

func TestChatGetCachedContacts(t *testing.T) {
	assert := assert.New(t)

	chat := new(Chat)
	chat.cacheContact(1, 2)
	assert.NotEmpty(chat.Contacts)
	assert.NotZero(chat.Contacts[1].Len())
	assert.True(chat.Contacts[1].Contains(2))
	assert.NotZero(chat.Contacts[2].Len())
	assert.True(chat.Contacts[2].Contains(1))

	contacts := chat.getCachedContacts(1)
	assert.Len(contacts, 1)
	assert.Equal(2, contacts[0])

	contacts = chat.getCachedContacts(2)
	assert.Len(contacts, 1)
	assert.Equal(1, contacts[0])
}

func TestChatRemoveCachedContacts(t *testing.T) {
	assert := assert.New(t)

	chat := new(Chat)
	chat.cacheContact(1, 2)
	assert.NotEmpty(chat.Contacts)
	assert.NotZero(chat.Contacts[1].Len())
	assert.True(chat.Contacts[1].Contains(2))
	assert.NotZero(chat.Contacts[2].Len())
	assert.True(chat.Contacts[2].Contains(1))

	chat.removeCachedContacts(1, 2)
	assert.Empty(chat.Contacts)

	chat.cacheContact(1, 2)
	assert.NotEmpty(chat.Contacts)
}

func TestChatGetCachedMessagesAfter(t *testing.T) {
	assert := assert.New(t)

	session1 := &model.Session{
		UUID:       "test_session",
		CreatedUTC: time.Now().UTC(),
		UserID:     1,
		User:       &model.User{ID: 1, UUID: "test_user1"},
	}
	session2 := &model.Session{
		UUID:       "test_session2",
		CreatedUTC: time.Now().UTC(),
		UserID:     2,
		User:       &model.User{ID: 2, UUID: "test_user2"},
	}

	chat := new(Chat)
	chat.cacheSession(session1)
	chat.addMessageQueue(session1)
	chat.cacheSession(session2)
	chat.addMessageQueue(session2)

	now := time.Now()
	chat.queueMessage(&model.Message{
		CreatedUTC: now.Add(-15 * time.Second),
		SenderID:   1,
		ReceiverID: 2,
		Body:       "This is a test message.",
	})
	chat.queueMessage(&model.Message{
		CreatedUTC: now.Add(-10 * time.Second),
		SenderID:   1,
		ReceiverID: 2,
		Body:       "This is a test message.",
	})
	chat.queueMessage(&model.Message{
		CreatedUTC: now.Add(-5 * time.Second),
		SenderID:   1,
		ReceiverID: 2,
		Body:       "This is a test message.",
	})
	chat.queueMessage(&model.Message{
		CreatedUTC: now.Add(-time.Second),
		SenderID:   1,
		ReceiverID: 2,
		Body:       "This is a test message.",
	})

	messages := chat.getCachedMessagesAfter(1, now.Add(-11*time.Second))
	assert.Len(messages, 3)

	messages = chat.getCachedMessagesAfter(2, now.Add(-11*time.Second))
	assert.Len(messages, 3)

	messages = chat.getCachedMessagesAfter(1, now)
	assert.Len(messages, 0)

	messages = chat.getCachedMessagesAfter(2, now)
	assert.Len(messages, 0)
}

func TestChatNewUserAction(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.IsolateTo(tx)
	chat := new(Chat)
	app.Register(chat)

	newUser := model.User{UUID: "a_test_user", DisplayName: "A Test user"}
	var response serviceResponseOfUser
	err = app.Mock().WithPathf("/api/user").WithVerb("POST").WithPostBodyAsJSON(&newUser).JSON(&response)
	assert.Nil(err)
	assert.False(response.Response.IsZero())
	assert.NotEmpty(chat.Users)
	assert.True(chat.hasCachedUser(response.Response.ID))
}

func TestChatGetUserAction(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u1, tx))

	app := web.New()
	app.IsolateTo(tx)
	chat := new(Chat)
	app.Register(chat)

	var response serviceResponseOfUser
	err = app.Mock().WithPathf("/api/user/%d", u1.ID).WithVerb("GET").JSON(&response)
	assert.Nil(err)
	assert.False(response.Response.IsZero())
	assert.Equal(u1.ID, response.Response.ID)
	assert.NotEmpty(chat.Users)
	assert.True(chat.hasCachedUser(response.Response.ID))
}

func TestChatUpdateUserAction(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u1, tx))

	app := web.New()
	app.IsolateTo(tx)
	chat := new(Chat)
	app.Register(chat)

	u2 := &model.User{UUID: u1.UUID, DisplayName: "Not Test User"}
	var response serviceResponseOfUser
	err = app.Mock().WithPathf("/api/user/%d", u1.ID).WithPostBodyAsJSON(u2).WithVerb("PUT").JSON(&response)
	assert.Nil(err)
	assert.Equal(http.StatusOK, response.Meta["http_code"], response.Meta["exception"])
	assert.Equal(u1.ID, response.Response.ID)

	var verify model.User
	err = model.DB().GetByIDInTransaction(&verify, tx, u1.ID)
	assert.Equal("Not Test User", verify.DisplayName)
}

func TestChatDeleteUserAction(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u1, tx))

	app := web.New()
	app.IsolateTo(tx)
	chat := new(Chat)
	app.Register(chat)

	meta, err := app.Mock().WithPathf("/api/user/%d", u1.ID).WithVerb("DELETE").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)

	var verify model.User
	err = model.DB().GetByIDInTransaction(&verify, tx, u1.ID)
	assert.True(verify.IsZero())
}

func TestSessionActions(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u1, tx))

	app := web.New()
	app.IsolateTo(tx)
	chat := new(Chat)
	app.Register(chat)

	var response serviceResponseOfSession
	err = app.Mock().WithVerb("POST").WithPathf("/api/session/%d", u1.ID).JSON(&response)
	assert.Nil(err)
	assert.Equal(http.StatusOK, response.Meta["http_code"])
	assert.False(response.Response.IsZero())
	assert.NotEmpty(chat.Sessions)
	assert.True(chat.userHasSession(u1.ID))

	var many serviceResponseOfSessions
	err = app.Mock().WithVerb("GET").WithPathf("/api/sessions").JSON(&many)
	assert.Nil(err)
	assert.NotEmpty(many.Response)

	meta, err := app.Mock().WithVerb("DELETE").WithPathf("/api/session/%s", response.Response.UUID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.False(chat.userHasSession(u1.ID))
}

func TestCreateContact(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u1, tx))

	u2 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u2, tx))

	s1 := &model.Session{
		UUID:          util.UUIDv4().ToShortString(),
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        u1.ID,
		User:          u1,
	}
	assert.Nil(model.DB().CreateInTransaction(s1, tx))

	app := web.New()
	app.IsolateTo(tx)
	chat := new(Chat)
	err = chat.Restore(tx)
	assert.Nil(err)
	app.Register(chat)

	meta, err := app.Mock().WithVerb("POST").WithPathf("/api/contact/%s/%d", s1.UUID, u2.ID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)

	var verify model.Contacts
	err = model.DB().QueryInTransaction("select * from contacts where sender = $1", tx, u1.ID).Out(&verify)
	assert.Equal(u1.ID, verify.Sender)
	assert.Equal(u2.ID, verify.Receiver)

	meta, err = app.Mock().WithVerb("POST").WithPathf("/api/contact/%s/%d", s1.UUID, u2.ID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
}

func TestSendMethod(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u1, tx))

	u2 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u2, tx))

	s1 := &model.Session{
		UUID:          "test_session",
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        u1.ID,
		User:          u1,
	}
	s2 := &model.Session{
		UUID:          "test_session2",
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        u2.ID,
		User:          u2,
	}
	assert.Nil(model.DB().CreateInTransaction(s1, tx))
	assert.Nil(model.DB().CreateInTransaction(s2, tx))

	app := web.New()
	app.IsolateTo(tx)
	chat := new(Chat)
	err = chat.Restore(tx)
	assert.Nil(err)
	app.Register(chat)

	message := model.Message{
		ReceiverID: u2.ID,
		Body:       "this is a test",
	}

	var response serviceResponseOfMessage
	err = app.Mock().WithVerb("POST").WithPathf("/api/message/%s", s1.UUID).WithPostBodyAsJSON(message).JSON(&response)
	assert.Nil(err)

	assert.NotEmpty(chat.MessageQueues)
	messages := chat.getCachedMessagesAfter(u1.ID, time.Now().UTC().Add(-time.Hour))
	assert.Len(messages, 1)
}

func TestGetMessages(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u1, tx))

	u2 := &model.User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(model.DB().CreateInTransaction(u2, tx))

	s1 := &model.Session{
		UUID:          "test_session",
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        u1.ID,
		User:          u1,
	}
	s2 := &model.Session{
		UUID:          "test_session2",
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        u2.ID,
		User:          u2,
	}
	assert.Nil(model.DB().CreateInTransaction(s1, tx))
	assert.Nil(model.DB().CreateInTransaction(s2, tx))

	message2 := model.Message{
		UUID:       util.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC().Add(-2 * time.Hour),
		SenderID:   u1.ID,
		ReceiverID: u2.ID,
		Body:       "this is a test",
	}
	assert.Nil(model.DB().CreateInTransaction(message2, tx))

	message3 := model.Message{
		UUID:       util.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC().Add(-time.Minute),
		SenderID:   u1.ID,
		ReceiverID: u2.ID,
		Body:       "this is a test",
	}
	assert.Nil(model.DB().CreateInTransaction(message3, tx))

	message := model.Message{
		UUID:       util.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC(),
		SenderID:   u1.ID,
		ReceiverID: u2.ID,
		Body:       "this is a test",
	}
	assert.Nil(model.DB().CreateInTransaction(message, tx))

	var messagesForU1 []model.Message
	err = model.DB().QueryInTransaction("select * from messages where sender = $1 or receiver = $1", tx, u1.ID).OutMany(&messagesForU1)
	assert.Nil(err)
	assert.Len(messagesForU1, 3)

	app := web.New()
	app.IsolateTo(tx)
	chat := new(Chat)
	err = chat.Restore(tx)
	assert.Nil(err)

	assert.Equal(3, chat.MessageQueues[u1.ID].Len())
	assert.Equal(3, chat.MessageQueues[u2.ID].Len())

	app.Register(chat)

	assert.NotEmpty(chat.MessageQueues)
	assert.NotZero(chat.MessageQueues[u1.ID].Len())

	cutoff := time.Now().UTC().Add(-time.Hour)
	assert.True(cutoff.After(message2.CreatedUTC))
	assert.True(cutoff.Before(message.CreatedUTC))

	var response serviceResponseOfMessages
	err = app.Mock().WithVerb("GET").WithPathf("/api/messages/%s/%d", s1.UUID, cutoff.Unix()).JSON(&response)
	assert.Nil(err)
	assert.Equal(http.StatusOK, response.Meta["http_code"])
	assert.Len(response.Response, 2)

	assert.True(response.Response[0].CreatedUTC.Before(response.Response[1].CreatedUTC))
}
