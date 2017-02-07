package server

import (
	"strings"

	"github.com/blendlabs/chatbus/server/controller"
	chronometer "github.com/blendlabs/go-chronometer"
	workQueue "github.com/blendlabs/go-workqueue"
	web "github.com/wcharczuk/go-web"
)

// optionsHandler totally defeats CORS.
func optionsHandler(rc *web.RequestContext) web.ControllerResult {
	rc.Response.Header().Set("Access-Control-Allow-Origin", rc.Request.Header.Get("Origin"))
	rc.Response.Header().Set("Access-Control-Allow-Methods", strings.ToUpper(rc.Request.Header.Get("Access-Control-Request-Method")))
	rc.Response.Header().Set("Access-Control-Allow-Headers", rc.Request.Header.Get("Access-Control-Request-Headers"))

	return rc.NoContent()
}

// New inits the http server.
func New() (*web.App, error) {
	app := web.New()
	app.SetName(DefaultConfig().AppName)
	app.SetPort(DefaultConfig().Port)
	//app.SetLogger(web.NewStandardOutputLogger())

	app.GET("/", func(context *web.RequestContext) web.ControllerResult {
		return context.API().JSON(map[string]interface{}{
			"app": DefaultConfig().AppName,
		})
	}, web.APIProviderAsDefault)

	// we have to do the following to allow the frontend to talk to this instance.
	app.OPTIONS("/*filepath", optionsHandler)
	app.RequestStartHandler(func(rc *web.RequestContext) {
		rc.Response.Header().Set("Access-Control-Allow-Origin", "*")
	})

	chatController := new(controller.Chat)
	err := chatController.Restore()
	if err != nil {
		return nil, err
	}
	app.Register(chatController)

	app.OnStart(func(app *web.App) error {
		chronometer.Default().LoadJob(&controller.CullSessions{Controller: chatController})
		chronometer.Default().Start()
		workQueue.Start(2)
		web.NewStandardOutputLogger().Log("Server started.")
		return nil
	})
	return app, nil
}
