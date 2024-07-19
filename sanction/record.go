package sanction

import (
	"context"
	"errors"
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

const (
	inComingRecordUpdateKey = "record:in_coming_update"
	redisSeparator          = ";"
)

var (
	BanCommandSanction Type
)

type ModRecord struct {
	gorm.Model
	UserID   string
	GuildID  string
	Proof    string
	Duration uint
	Reason   string
	Sanction *Type
	Removed  bool `gorm:"default:false"`
}

type Member struct {
	UserID  string
	GuildID string
}

func (m *Member) GetAllRecord() ([]*ModRecord, error) {
	var records []*ModRecord
	err := gokord.DB.Preload("Sanctions").Where("user_id = ?", m.UserID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (m *Member) Sanction(record *ModRecord) error {
	record.UserID = m.UserID
	record.GuildID = m.GuildID
	err := gokord.DB.Create(record).Error
	if err != nil {
		return err
	}
	return nil
}

func CreateSanction(sanction *Type, reason string, proof string, duration uint) *ModRecord {
	return &ModRecord{
		Reason:   reason,
		Proof:    proof,
		Duration: duration,
		Sanction: sanction,
	}
}

func UpdateRecord(s *discordgo.Session, client *redis.Client) {
	v := client.Get(context.Background(), inComingRecordUpdateKey)
	var id []int
	if v.Err() != nil && !errors.Is(v.Err(), redis.Nil) {
		utils.SendAlert("record.go - fetching in coming update", v.Err().Error())
	} else if v.Err() == nil {
		splitted := strings.Split(v.Val(), redisSeparator)
		for _, s := range splitted {
			val, err := strconv.Atoi(s)
			if err != nil {
				utils.SendAlert("record.go - parsing record id", err.Error(), "val", val)
				continue
			}
			id = append(id, val)
		}
		var records []*ModRecord
		gokord.DB.Preload("Types").Find(&records, id)
		for _, r := range records {
			if r.CreatedAt.Unix()+int64(r.Duration) <= time.Now().Unix() && !r.Removed {
				if r.Sanction.ID == BanCommandSanction.ID {
					err := s.GuildBanDelete(r.GuildID, r.UserID)
					if err != nil {
						utils.SendAlert(
							"record.go - unbanning",
							err.Error(),
							"guild", r.GuildID,
							"user", r.UserID,
							"record", r.ID,
						)
					}
				} else {
					utils.SendWarn("Not a ban", "guild", r.GuildID, "user", r.UserID, "record", r.ID)
				}
				r.Removed = true
				gokord.DB.Save(&r)
				continue
			}
			ret, ok := retentions[r.Sanction.Retention]
			if !ok {
				utils.SendAlert(
					"record.go - retention not found",
					"",
					"guild", r.GuildID,
					"user", r.UserID,
					"record", r.ID,
					"retention", r.Sanction.Retention,
				)
				continue
			}
			if !((ret.RelativeRetention == 0 && r.CreatedAt.Unix()+ret.Retention <= time.Now().Unix()) ||
				(ret.Retention == 0 && r.CreatedAt.Unix()*int64(ret.RelativeRetention) <= time.Now().Unix())) {
				continue
			}
			err := gokord.DB.Delete(&r).Error
			if err != nil {
				utils.SendAlert(
					"record.go - deleting record",
					err.Error(),
					"guild", r.GuildID,
					"user", r.UserID,
					"record", r.ID,
				)
				continue
			}
			n := id
			for i, v := range id {
				if v == int(r.ID) {
					n = append(id[:i], id[i+1:]...)
				}
			}
			id = n
		}
		joined := ""
		for _, i := range id {
			joined += fmt.Sprintf("%d%s", i, redisSeparator)
		}
		joined = strings.TrimSuffix(joined, redisSeparator)
		err := client.Set(context.Background(), inComingRecordUpdateKey, joined, 0).Err()
		if err != nil {
			utils.SendAlert("record.go - setting in coming record update key 1", err.Error(), "joined", joined)
		}
	}
	var records []*ModRecord
	var err error
	if len(id) != 0 {
		err = gokord.DB.Where("id NOT IN ?", id).Find(&records).Error
	} else {
		err = gokord.DB.Find(&records).Error
	}
	if err != nil {
		utils.SendAlert("record.go - fetching all records", err.Error(), "ids", id)
	}
	var add []int
	for _, r := range records {
		if r.CreatedAt.Unix()+int64(r.Duration) >= time.Now().Unix() {
			add = append(add, int(r.ID))
		}
	}
	v = client.Get(context.Background(), inComingRecordUpdateKey)
	if v.Err() != nil {
		utils.SendAlert("record.go - getting records in coming update", v.Err().Error())
		return
	}
	joined := v.Val()
	for _, i := range add {
		joined += fmt.Sprintf("%d%s", i, redisSeparator)
	}
	joined = strings.TrimSuffix(joined, redisSeparator)
	err = client.Set(context.Background(), inComingRecordUpdateKey, joined, 0).Err()
	if err != nil {
		utils.SendAlert("record.go - setting in coming record update key 2", err.Error(), "joined", joined)
	}
}
