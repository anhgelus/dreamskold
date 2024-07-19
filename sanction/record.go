package sanction

import (
	"database/sql"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
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
}

type Member struct {
	UserID  string
	GuildID string
}

func (m *Member) GetAllRecord() ([]*ModRecord, error) {
	res := gokord.DB.Preload("Sanctions").Where("user_id = ?", m.UserID).Find(&ModRecord{})
	if res.Error != nil {
		return nil, res.Error
	}
	rows, err := res.Rows()
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			utils.SendAlert("record.go - closing rows", err.Error())
		}
	}(rows)
	var records []*ModRecord
	for rows.Next() {
		var record ModRecord
		err = rows.Scan(&record)
		if err != nil {
			utils.SendAlert("record.go - scanning record", err.Error())
			continue
		}
		records = append(records, &record)
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

func UpdateRecord(s *discordgo.Session) {
	// query all record in redis and check (in goroutine)
	// query all record except removed one and stored one in redis
	//	if they expire in less than one day, put them in redis
}
