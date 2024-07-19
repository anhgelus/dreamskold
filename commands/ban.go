package commands

import (
	"fmt"
	"github.com/anhgelus/dreamskold/sanction"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func Ban(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := utils.ResponseBuilder{C: s, I: i}
	optMap := utils.GenerateOptionMap(i)
	v, ok := optMap["member"]
	if !ok {
		err := resp.IsEphemeral().Message("Member is not set.").Send()
		if err != nil {
			utils.SendAlert("ban.go - send error member is not set", err.Error())
		}
		return
	}
	member := v.UserValue(s)
	v, ok = optMap["duration"]
	if !ok {
		err := resp.IsEphemeral().Message("Duration is not set.").Send()
		if err != nil {
			utils.SendAlert("ban.go - send error duration is not set", err.Error())
		}
		return
	}
	duration := v.StringValue()
	v, ok = optMap["reason"]
	if !ok {
		err := resp.IsEphemeral().Message("Reason is not set.").Send()
		if err != nil {
			utils.SendAlert("ban.go - send error reason reason is not set", err.Error())
		}
		return
	}
	reason := v.StringValue()
	v, ok = optMap["proof"]
	if !ok {
		err := resp.IsEphemeral().Message("Proof is not set.").Send()
		if err != nil {
			utils.SendAlert("ban.go - send error proof reason is not set", err.Error())
		}
		return
	}
	proof := v.StringValue()
	proofIDs := strings.Split(proof[len(proof)-38:], "/")
	proofChannelID := proofIDs[0]
	proofMessageID := proofIDs[1]
	messageProof, err := s.ChannelMessage(proofChannelID, proofMessageID)
	if err != nil {
		utils.SendAlert("ban.go - failed to get message proof", err.Error())
		err = resp.IsEphemeral().Message("Failed to get message proof").Send()
		if err != nil {
			utils.SendAlert("ban.go - send error failed to get message proof", err.Error())
		}
		return
	}
	if messageProof == nil {
		utils.SendAlert("ban.go - failed to get message proof", "message proof is nil")
		err = resp.IsEphemeral().Message("Failed to get message proof").Send()
		if err != nil {
			utils.SendAlert("ban.go - send error failed to get message proof", err.Error())
		}
		return
	}

	m := sanction.Member{GuildID: i.GuildID, UserID: member.ID}
	dur, err := sanction.LoadDuration(duration)
	if err != nil {
		utils.SendAlert("ban.go - failed to load duration", err.Error())
		err = resp.IsEphemeral().Message("Failed to load duration: " + err.Error()).Send()
		if err != nil {
			utils.SendAlert("ban.go - send error failed to load duration", err.Error())
		}
		return
	}
	err = m.Sanction(sanction.CreateSanction(nil, reason, messageProof.Content, dur.ToUint()))
	if err != nil {
		utils.SendAlert("ban.go - failed to apply sanction", err.Error())
		err = resp.IsEphemeral().Message("Failed to apply sanction").Send()
		if err != nil {
			utils.SendAlert("ban.go - send error failed to apply sanction", err.Error())
		}
		return
	}
	_, err = s.GuildBan(i.GuildID, member.ID)
	if err != nil {
		utils.SendAlert("ban.go - failed to ban member", err.Error(), "username", member.Username)
		err = resp.IsEphemeral().Message("Failed to ban member").Send()
		if err != nil {
			utils.SendAlert("ban.go - send error failed to ban member", err.Error())
		}
		return
	}
	err = resp.Message(fmt.Sprintf("%s banned for %s", member.Mention(), "")).Send()
	if err != nil {
		utils.SendAlert("ban.go - send success", err.Error())
	}
}
