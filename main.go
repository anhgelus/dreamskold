package main

import (
	_ "embed"
	"flag"
	"github.com/anhgelus/dreamskold/commands"
	"github.com/anhgelus/dreamskold/config"
	"github.com/anhgelus/dreamskold/sanction"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"time"
)

var (
	token string
	//go:embed updates.json
	updatesData []byte
	Version     = gokord.Version{
		Major: 0,
		Minor: 0,
		Patch: 0,
	} // git version: 0.0.0
	sanctionConfig config.SanctionConfig
)

func init() {
	flag.StringVar(&token, "token", "", "token of the bot")
	flag.Parse()
}

func main() {
	err := gokord.SetupConfigs([]*gokord.ConfigInfo{
		{
			Cfg:     &sanctionConfig,
			Name:    "sanction",
			Default: config.DefaultSanctionConfig,
		},
	})
	if err != nil {
		panic(err)
	}

	err = sanction.SetupSanctionTypes(sanctionConfig.Sanctions, sanctionConfig.Retentions)
	if err != nil {
		panic(err)
	}

	err = gokord.DB.AutoMigrate(&sanction.Type{}, &sanction.ModRecord{})
	if err != nil {
		panic(err)
	}

	banSanctionType := sanction.Type{Name: sanctionConfig.BanCommandSanction}
	err = gokord.DB.First(&banSanctionType).Error
	if err != nil {
		panic(err)
	}
	sanction.BanCommandSanction = banSanctionType

	innovations, err := gokord.LoadInnovationFromJson(updatesData)
	if err != nil {
		panic(err)
	}

	adm := gokord.AdminPermission
	ban := int64(discordgo.PermissionBanMembers)
	mod := int64(discordgo.PermissionModerateMembers)

	banCmd := gokord.NewCommand("ban", "Ban a member").
		HasOption().
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"member",
			"Member to ban",
		).IsRequired()).
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionString,
			"duration",
			"Duration of the ban",
		).IsRequired()).
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionString,
			"reason",
			"Reason of the ban",
		).IsRequired()).
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionString,
			"proof",
			"Link to the proof (discord message)",
		).IsRequired()).
		SetPermission(&ban).
		SetHandler(commands.Ban)

	bot := gokord.Bot{
		Token: token,
		Status: []*gokord.Status{
			{
				Type:    gokord.WatchStatus,
				Content: "DreamSköld",
			},
			{
				Type:    gokord.GameStatus,
				Content: "être dev par @anhgelus",
			},
			{
				Type:    gokord.ListeningStatus,
				Content: "http 418, I'm a tea pot",
			},
			{
				Type:    gokord.GameStatus,
				Content: "DreamSköld " + Version.String(),
			},
		},
		Commands: []*gokord.GeneralCommand{
			banCmd,
		},
		AfterInit:   afterInit,
		Version:     nil,
		Innovations: innovations,
		Name:        "DreamSköld",
	}
	bot.Start()
}

func afterInit(dg *discordgo.Session) {
	client, err := gokord.BaseCfg.Redis.Get()
	if err != nil {
		panic(err)
	}

	utils.NewTimer(15*time.Minute, func(stop chan struct{}) {
		sanction.UpdateRecord(dg, client)
	})
}
