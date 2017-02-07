package controller

import (
	"database/sql"
	"strconv"
	"sync"
	"time"

	"github.com/blendlabs/chatbus/server/model"
	"github.com/blendlabs/chatbus/server/viewmodel"
	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/collections"
	web "github.com/wcharczuk/go-web"
)

const (
	// MessageQueueMaxLength is the maximum queue length per user.
	MessageQueueMaxLength = 1 << 10 //1 << 18 // 256k
)

// Chat is the chat controller
type Chat struct {
	usersLock         sync.RWMutex
	contactsLock      sync.RWMutex
	sessionLock       sync.RWMutex
	sessionByUserLock sync.RWMutex
	messageQueueLock  sync.RWMutex

	App *web.App

	Users          map[int]*model.User
	Contacts       map[int]collections.SetOfInt
	Sessions       map[string]*model.Session
	SessionsByUser map[int]collections.SetOfString
	MessageQueues  map[int]*collections.RingBuffer
}

// Register registers the controller.
func (c *Chat) Register(app *web.App) {
	// save an app reference.
	c.App = app

	// user actions
	app.GET("/api/users", c.getUsersAction, web.APIProviderAsDefault)
	app.POST("/api/user", c.newUserAction, web.APIProviderAsDefault)
	app.GET("/api/user/:id", c.getUserAction, web.APIProviderAsDefault)
	app.GET("/api/user.uuid/:uuid", c.getUserByUUIDAction, web.APIProviderAsDefault)
	app.PUT("/api/user/:id", c.updateUserAction, web.APIProviderAsDefault)
	app.DELETE("/api/user/:id", c.deleteUserAction, web.APIProviderAsDefault)

	// session actions
	app.GET("/api/sessions", c.getSessionsAction, web.APIProviderAsDefault)
	app.POST("/api/session/:user_id", c.newSessionAction, web.APIProviderAsDefault)
	app.DELETE("/api/session/:id", c.deleteSessionAction, web.APIProviderAsDefault)

	// contacts actions
	app.GET("/api/contacts/:session_id", c.getContactsAction, web.APIProviderAsDefault)
	app.POST("/api/contact/:session_id/:user_id", c.createContactAction, web.APIProviderAsDefault)
	app.DELETE("/api/contact/:session_id/:user_id", c.deleteContactAction, web.APIProviderAsDefault)

	// messages actions
	app.GET("/api/messages/:session_id", c.getMessagesAction, web.APIProviderAsDefault)
	app.GET("/api/messages/:session_id/:after", c.getMessagesAction, web.APIProviderAsDefault)
	app.GET("/api/messages/:session_id/:after/:nano", c.getMessagesAction, web.APIProviderAsDefault)
	app.POST("/api/message/:session_id", c.sendMessageAction, web.APIProviderAsDefault)
}

// Restore restores the chat controller from state in the db
func (c *Chat) Restore(txs ...*sql.Tx) error {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}
	var users []model.User
	err := model.DB().GetAllInTransaction(&users, tx)
	if err != nil {
		return err
	}
	for x := 0; x < len(users); x++ {
		user := users[x]
		c.cacheUser(&user)
	}

	var sessions []model.Session
	err = model.DB().GetAllInTransaction(&sessions, tx)
	if err != nil {
		return err
	}

	for x := 0; x < len(sessions); x++ {
		session := sessions[x]
		session.User = c.getCachedUser(session.UserID)

		c.addMessageQueue(&session)
		c.cacheSession(&session)
		c.cacheSessionByUser(&session)
	}

	var contacts []model.Contacts
	err = model.DB().GetAllInTransaction(&contacts, tx)
	if err != nil {
		return err
	}
	for x := 0; x < len(contacts); x++ {
		contact := contacts[x]
		c.cacheContact(contact.Sender, contact.Receiver)
	}

	messages, err := model.GetAllMessagesWithLimit(MessageQueueMaxLength, tx)
	if err != nil {
		return err
	}

	for x := 0; x < len(messages); x++ {
		message := messages[x]
		c.queueMessage(&message)
	}

	return nil
}

func (c *Chat) cacheUser(user *model.User) {
	c.usersLock.Lock()
	defer c.usersLock.Unlock()
	if c.Users == nil {
		c.Users = map[int]*model.User{}
	}
	c.Users[user.ID] = user
}

func (c *Chat) cacheContact(sender, receiver int) {
	c.contactsLock.Lock()
	defer c.contactsLock.Unlock()
	if c.Contacts == nil {
		c.Contacts = map[int]collections.SetOfInt{}
	}
	if _, hasContacts := c.Contacts[sender]; !hasContacts {
		c.Contacts[sender] = collections.NewSetOfInt()
	}
	c.Contacts[sender].Add(receiver)

	if _, hasContacts := c.Contacts[receiver]; !hasContacts {
		c.Contacts[receiver] = collections.NewSetOfInt()
	}
	c.Contacts[receiver].Add(sender)
}

func (c *Chat) cacheSession(session *model.Session) {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()
	if c.Sessions == nil {
		c.Sessions = map[string]*model.Session{}
	}
	c.Sessions[session.UUID] = session
}

func (c *Chat) cacheSessionByUser(session *model.Session) {
	c.sessionByUserLock.Lock()
	defer c.sessionByUserLock.Unlock()
	if c.SessionsByUser == nil {
		c.SessionsByUser = map[int]collections.SetOfString{}
	}
	if _, hasSessions := c.SessionsByUser[session.UserID]; !hasSessions {
		c.SessionsByUser[session.UserID] = collections.NewSetOfString()
	}
	c.SessionsByUser[session.UserID].Add(session.UUID)
}

func (c *Chat) userHasSession(userID int) bool {
	c.sessionByUserLock.RLock()
	defer c.sessionByUserLock.RUnlock()
	if _, hasSessions := c.SessionsByUser[userID]; hasSessions {
		return true
	}
	return false
}

func (c *Chat) addMessageQueue(session *model.Session) {
	c.messageQueueLock.Lock()
	defer c.messageQueueLock.Unlock()

	if c.MessageQueues == nil {
		c.MessageQueues = map[int]*collections.RingBuffer{}
	}

	c.MessageQueues[session.UserID] = collections.NewRingBufferWithCapacity(1024)
}

func (c *Chat) setCachedSessionLastActive(sessionID string) {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()
	c.Sessions[sessionID].LastActiveUTC = time.Now().UTC()
}

func (c *Chat) getCachedUser(userID int) *model.User {
	c.usersLock.RLock()
	defer c.usersLock.RUnlock()

	if user, hasUser := c.Users[userID]; hasUser {
		return user
	}

	return nil
}

func (c *Chat) hasCachedUser(userID int) bool {
	c.usersLock.RLock()
	defer c.usersLock.RUnlock()
	if _, hasUser := c.Users[userID]; hasUser {
		return true
	}
	return false
}

func (c *Chat) queueMessage(message *model.Message) {
	c.messageQueueLock.RLock()
	defer c.messageQueueLock.RUnlock()

	if c.MessageQueues == nil {
		return
	}

	if queue, hasQueue := c.MessageQueues[message.SenderID]; hasQueue {
		func() {
			queue.SyncRoot().Lock()
			defer queue.SyncRoot().Unlock()
			if queue.Len() >= MessageQueueMaxLength {
				queue.Dequeue()
			}
			queue.Enqueue(message)
		}()
	}

	if queue, hasQueue := c.MessageQueues[message.ReceiverID]; hasQueue {
		func() {
			queue.SyncRoot().Lock()
			defer queue.SyncRoot().Unlock()

			if queue.Len() >= MessageQueueMaxLength {
				queue.Dequeue()
			}
			queue.Enqueue(message)
		}()
	}
}

func (c *Chat) removeCachedUser(userID int) {
	c.usersLock.Lock()
	defer c.usersLock.Unlock()
	delete(c.Users, userID)
}

func (c *Chat) getCachedSession(sessionID string) (*model.Session, bool) {
	c.sessionLock.RLock()
	defer c.sessionLock.RUnlock()
	if session, hasSession := c.Sessions[sessionID]; hasSession {
		return session, hasSession
	}
	return nil, false
}

func (c *Chat) removeCachedSession(sessionID string) {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()
	delete(c.Sessions, sessionID)
}

func (c *Chat) removeCachedSessionByUser(session *model.Session) {
	c.sessionByUserLock.Lock()
	defer c.sessionByUserLock.Unlock()

	if sessionSet, hasSessions := c.SessionsByUser[session.UserID]; hasSessions {
		sessionSet.Remove(session.UUID)
		if sessionSet.Len() == 0 {
			delete(c.SessionsByUser, session.UserID)
		}
	}
}

func (c *Chat) getCachedContacts(sender int) []int {
	c.contactsLock.RLock()
	defer c.contactsLock.RUnlock()
	if contacts, hasContacts := c.Contacts[sender]; hasContacts {
		return contacts.AsSlice()
	}
	return []int{}
}

func (c *Chat) removeCachedContacts(sender, receiver int) {
	c.contactsLock.Lock()
	defer c.contactsLock.Unlock()
	if _, hasContact := c.Contacts[sender]; hasContact {
		c.Contacts[sender].Remove(receiver)
	}
	if c.Contacts[sender].Len() == 0 {
		delete(c.Contacts, sender)
	}
	if _, hasContact := c.Contacts[receiver]; hasContact {
		c.Contacts[receiver].Remove(sender)
	}
	if c.Contacts[receiver].Len() == 0 {
		delete(c.Contacts, receiver)
	}
}

func (c *Chat) getCachedMessagesAfter(userID int, cutoff time.Time) []model.Message {
	c.messageQueueLock.RLock()
	defer c.messageQueueLock.RUnlock()

	messages := []model.Message{}
	if queue, hasQueue := c.MessageQueues[userID]; hasQueue {
		queue.ReverseEachUntil(func(v interface{}) bool {
			message := model.TryCastMessage(v)
			if message.CreatedUTC.After(cutoff) {
				messages = append(messages, *message)
				return true
			}
			return false
		})
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages
}

// GET /api/users
func (c *Chat) getUsersAction(rc *web.RequestContext) web.ControllerResult {
	users := []model.User{}
	err := model.DB().GetAllInTransaction(&users, rc.Tx())
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().JSON(users)
}

// POST /api/users
func (c *Chat) newUserAction(rc *web.RequestContext) web.ControllerResult {
	var user model.User
	err := rc.PostBodyAsJSON(&user)
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	existingUser, err := model.GetUserByUUID(user.UUID)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if !existingUser.IsZero() {
		c.cacheUser(existingUser)
		return rc.API().JSON(existingUser)
	}

	err = model.DB().CreateInTransaction(&user, rc.Tx())
	if err != nil {
		return rc.API().InternalError(err)
	}
	c.cacheUser(&user)
	return rc.API().JSON(user)
}

// GET /api/user/:id
func (c *Chat) getUserAction(rc *web.RequestContext) web.ControllerResult {
	userID, err := rc.RouteParameterInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	var user model.User
	err = model.DB().GetByIDInTransaction(&user, rc.Tx(), userID)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if user.IsZero() {
		return rc.API().NotFound()
	}
	c.cacheUser(&user)
	return rc.API().JSON(user)
}

// GET /api/user.uuid/:uuid
func (c *Chat) getUserByUUIDAction(rc *web.RequestContext) web.ControllerResult {
	uuid, err := rc.RouteParameter("uuid")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	user, err := model.GetUserByUUID(uuid, rc.Tx())
	if err != nil {
		return rc.API().InternalError(err)
	}
	if user.IsZero() {
		return rc.API().NotFound()
	}
	c.cacheUser(user)
	return rc.API().JSON(user)
}

// PUT /api/user/:id
func (c *Chat) updateUserAction(rc *web.RequestContext) web.ControllerResult {
	userID, err := rc.RouteParameterInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	var user model.User
	err = model.DB().GetByIDInTransaction(&user, rc.Tx(), userID)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if user.IsZero() {
		return rc.API().NotFound()
	}

	var postedUser model.User
	err = rc.PostBodyAsJSON(&postedUser)
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	postedUser.ID = user.ID
	postedUser.UUID = user.UUID
	err = model.DB().UpdateInTransaction(&postedUser, rc.Tx())
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().JSON(user)
}

// DELETE /api/user/:id
func (c *Chat) deleteUserAction(rc *web.RequestContext) web.ControllerResult {
	userID, err := rc.RouteParameterInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	var user model.User
	err = model.DB().GetByIDInTransaction(&user, rc.Tx(), userID)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if user.IsZero() {
		return rc.API().NotFound()
	}
	err = model.DB().DeleteInTransaction(&user, rc.Tx())
	if err != nil {
		return rc.API().InternalError(err)
	}
	c.removeCachedUser(user.ID)
	return rc.API().OK()
}

// GET /api/sessions
func (c *Chat) getSessionsAction(rc *web.RequestContext) web.ControllerResult {
	c.sessionLock.RLock()
	defer c.sessionLock.RUnlock()

	output := []model.Session{}
	for _, session := range c.Sessions {
		output = append(output, *session)
	}
	return rc.API().JSON(output)
}

// POST /api/session
func (c *Chat) newSessionAction(rc *web.RequestContext) web.ControllerResult {
	userID, err := rc.RouteParameter("user_id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var user model.User
	err = model.DB().GetByIDInTransaction(&user, rc.Tx(), userID)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if user.IsZero() {
		return rc.API().NotFound()
	}

	newSession := &model.Session{
		UUID:          util.UUIDv4().ToShortString(),
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        user.ID,
		User:          &user,
	}

	err = model.DB().CreateInTransaction(newSession, rc.Tx())
	if err != nil {
		return rc.API().InternalError(err)
	}

	c.cacheUser(&user)
	c.cacheSession(newSession)
	c.cacheSessionByUser(newSession)
	c.addMessageQueue(newSession)
	return rc.API().JSON(newSession)
}

// DELETE /api/session/:id
func (c *Chat) deleteSessionAction(rc *web.RequestContext) web.ControllerResult {
	sessionID, err := rc.RouteParameter("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var session model.Session
	err = model.DB().GetByIDInTransaction(&session, rc.Tx(), sessionID)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if session.IsZero() {
		return rc.API().NotFound()
	}
	err = c.deleteSession(&session)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().OK()
}

func (c *Chat) deleteSession(session *model.Session, txs ...*sql.Tx) error {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}
	err := model.DB().DeleteInTransaction(session, tx)
	if err != nil {
		return err
	}
	c.removeCachedSession(session.UUID)
	c.removeCachedSessionByUser(session)
	return nil
}

// GET /api/contacts/:session_id
func (c *Chat) getContactsAction(rc *web.RequestContext) web.ControllerResult {
	sessionID, err := rc.RouteParameter("session_id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	session, hasSession := c.getCachedSession(sessionID)
	if !hasSession {
		return rc.API().NotFound()
	}
	contactIDs := c.getCachedContacts(session.UserID)

	output := []viewmodel.Contact{}
	for _, id := range contactIDs {
		isOnline := c.userHasSession(id)
		if user, hasUser := c.Users[id]; hasUser {
			output = append(output, viewmodel.Contact{
				User:     user,
				IsOnline: isOnline,
			})
		} else {
			var user model.User
			err = model.DB().GetByIDInTransaction(&user, rc.Tx(), id)
			if err != nil {
				return rc.API().InternalError(err)
			}
			output = append(output, viewmodel.Contact{
				User:     &user,
				IsOnline: isOnline,
			})
		}
	}
	return rc.API().JSON(output)
}

// POST /api/contacts/:session_id/:user_id
func (c *Chat) createContactAction(rc *web.RequestContext) web.ControllerResult {
	sessionID, err := rc.RouteParameter("session_id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	session, hasSession := c.getCachedSession(sessionID)
	if !hasSession {
		return rc.API().NotFound()
	}

	userID, err := rc.RouteParameterInt("user_id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	var user model.User
	err = model.DB().GetByIDInTransaction(&user, rc.Tx(), userID)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if user.IsZero() {
		return rc.API().NotFound()
	}

	c1 := model.Contacts{Sender: session.UserID, Receiver: user.ID}
	c2 := model.Contacts{Sender: user.ID, Receiver: session.UserID}

	if exists, _ := model.DB().ExistsInTransaction(c1, rc.Tx()); !exists {
		err = model.DB().CreateInTransaction(c1, rc.Tx())
		if err != nil {
			return rc.API().InternalError(err)
		}
	}

	if exists, _ := model.DB().ExistsInTransaction(c2, rc.Tx()); !exists {
		err = model.DB().CreateInTransaction(c2, rc.Tx())
		if err != nil {
			return rc.API().InternalError(err)
		}
	}

	c.cacheContact(session.UserID, userID)
	c.cacheContact(userID, session.UserID)
	return rc.API().OK()
}

// DELETE /api/contacts/:session_id/:user_id
func (c *Chat) deleteContactAction(rc *web.RequestContext) web.ControllerResult {
	sessionID, err := rc.RouteParameter("session_id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	session, hasSession := c.getCachedSession(sessionID)
	if !hasSession {
		return rc.API().NotFound()
	}

	userID, err := rc.RouteParameterInt("user_id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	var user model.User
	err = model.DB().GetByIDInTransaction(&user, rc.Tx(), userID)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if user.IsZero() {
		return rc.API().NotFound()
	}
	err = model.DeleteContacts(session.UserID, userID)
	if err != nil {
		return rc.API().InternalError(err)
	}

	c.removeCachedContacts(session.UserID, userID)
	return rc.API().OK()
}

// GET /api/messages/:id/:after
func (c *Chat) getMessagesAction(rc *web.RequestContext) web.ControllerResult {
	sessionID, err := rc.RouteParameter("session_id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	session, hasSession := c.getCachedSession(sessionID)
	if !hasSession {
		return rc.API().NotFound()
	}

	var after int64
	afterStr, _ := rc.RouteParameter("after")
	if len(afterStr) > 0 {
		after, err = strconv.ParseInt(afterStr, 10, 64)
		if err != nil {
			return rc.API().BadRequest(err.Error())
		}
	}

	var afterNano int64
	nanoStr, _ := rc.RouteParameter("nano")
	if len(nanoStr) > 0 {
		afterNano, err = strconv.ParseInt(nanoStr, 10, 64)
		if err != nil {
			return rc.API().BadRequest(err.Error())
		}
	}

	c.setCachedSessionLastActive(session.UUID)
	cutoff := time.Unix(after, afterNano).UTC()
	messages := c.getCachedMessagesAfter(session.UserID, cutoff)
	return rc.API().JSON(messages)
}

// POST /api/send/:id
func (c *Chat) sendMessageAction(rc *web.RequestContext) web.ControllerResult {
	sessionID, err := rc.RouteParameter("session_id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	session, hasSession := c.getCachedSession(sessionID)
	if !hasSession {
		return rc.API().NotFound()
	}

	var message model.Message
	err = rc.PostBodyAsJSON(&message)
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	if !c.hasCachedUser(message.ReceiverID) {
		return rc.API().BadRequest("Recipient not found!")
	}

	message.CreatedUTC = time.Now().UTC()
	message.SenderID = session.UserID
	message.UUID = util.UUIDv4().ToShortString()

	c.queueMessage(&message)
	c.setCachedSessionLastActive(session.UUID)

	message.QueueCreate()
	return rc.API().JSON(message)
}
