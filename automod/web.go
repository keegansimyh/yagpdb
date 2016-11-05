package automod

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jonas747/yagpdb/bot"
	"github.com/jonas747/yagpdb/web"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
	"html/template"
	"net/http"
)

type CtxKey int

const (
	CurrentConfig CtxKey = iota
)

type GeneralForm struct {
	Enabled bool
}

func (p *Plugin) InitWeb() {
	web.Templates = template.Must(web.Templates.ParseFiles("templates/plugins/automod.html"))

	autmodMux := goji.SubMux()
	web.CPMux.HandleC(pat.New("/automod/*"), autmodMux)
	web.CPMux.HandleC(pat.New("/automod"), autmodMux)

	// Alll handlers here require guild channels present
	autmodMux.UseC(web.RequireFullGuildMW)
	autmodMux.UseC(web.RequireGuildChannelsMiddleware)

	getHandler := web.RenderHandler(HandleAutomod, "cp_automod")

	autmodMux.HandleC(pat.Get("/"), getHandler)
	autmodMux.HandleC(pat.Get(""), getHandler)

	// Post handlers

	autmodMux.HandleC(pat.Post("/"), ExtraPostMW(web.SimpleConfigSaverHandler(Config{}, getHandler)))
	autmodMux.HandleC(pat.Post(""), ExtraPostMW(web.SimpleConfigSaverHandler(Config{}, getHandler)))
}

func HandleAutomod(ctx context.Context, w http.ResponseWriter, r *http.Request) interface{} {
	client, g, templateData := web.GetBaseCPContextData(ctx)

	config, err := GetConfig(client, g.ID)
	web.CheckErr(templateData, err, "Failed retrieving rules", log.Error)

	templateData["AutomodConfig"] = config
	templateData["VisibleURL"] = "/cp/" + g.ID + "/automod/"

	return templateData
}

// Invalidates the cache
func ExtraPostMW(inner goji.Handler) goji.Handler {
	mw := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		client, activeGuild, _ := web.GetBaseCPContextData(ctx)
		bot.PublishEvent(client, "update_automod_rules", activeGuild.ID, nil)
		inner.ServeHTTPC(ctx, w, r)
	}
	return goji.HandlerFunc(mw)
}