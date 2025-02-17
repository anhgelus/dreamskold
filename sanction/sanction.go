package sanction

import (
	"errors"
	"fmt"
	"github.com/anhgelus/dreamskold/config"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

type Type struct {
	gorm.Model
	Name        string `gorm:"unique" toml:"name"`
	Description string `toml:"description"`
	Retention   int8   `toml:"retention"`
}

type Duration struct {
	Days   uint
	Months uint
	Years  uint
}

var (
	types      []*Type
	retentions map[int8]*config.RetentionConfig
)

func SetupSanctionTypes(t []*Type, retCfg []*config.RetentionConfig) error {
	var scannedTypes []*Type
	err := gokord.DB.Find(&scannedTypes).Error
	if err != nil {
		return err
	}
	toDelete := scannedTypes
	for _, ty := range t {
		valid := false
		for i, sty := range scannedTypes {
			if ty.Name == sty.Name {
				valid = true
				toDelete = append(toDelete[:i], toDelete[i+1:]...)
			}
		}
		if !valid {
			err = gokord.DB.Create(ty).Error
			if err != nil {
				utils.SendAlert("sanction.go - creating new type", err.Error(), "name", ty.Name)
			}
		}
	}
	for _, sty := range toDelete {
		err = gokord.DB.Delete(sty).Error
		if err != nil {
			utils.SendAlert("sanction.go - deleting type", err.Error(), "name", sty.Name)
		}
	}
	types = t
	var retMap map[int8]*config.RetentionConfig
	for _, ret := range retCfg {
		retMap[ret.ID] = ret
	}
	retentions = retMap
	return nil
}

func LoadDuration(duration string) (*Duration, error) {
	splitted := strings.Split(duration, " ")
	if len(splitted) > 3 {
		return nil, errors.New("invalid duration")
	}
	dur := Duration{}
	for _, s := range splitted {
		runed := []rune(s)
		t := runed[len(runed)-1]
		v, err := strconv.Atoi(string(runed[:len(runed)-1]))
		if err != nil {
			return nil, err
		}
		switch t {
		case 'd':
			dur.Days = uint(v)
		case 'm':
			dur.Months = uint(v)
		case 'y':
			dur.Years = uint(v)
		default:
			return nil, errors.New("invalid duration")
		}
	}
	return &dur, nil
}

func (d *Duration) ToString() string {
	str := ""
	if d.Years > 0 {
		str += fmt.Sprintf("%d years", d.Years)
	}
	if d.Months > 0 {
		if str != "" {
			str += " "
		}
		str += fmt.Sprintf("%d months", d.Months)
	}
	if d.Days > 0 {
		if str != "" {
			str += " "
		}
		str += fmt.Sprintf("%d days", d.Days)
	}
	return str
}

func (d *Duration) ToUint() uint {
	day := 24 * uint(time.Hour)
	return d.Days*day + d.Months*30 + d.Years*12*30*day
}
