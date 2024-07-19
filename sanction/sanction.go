package sanction

import (
	"database/sql"
	"errors"
	"fmt"
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
	Retention   uint8  `toml:"retention"`
}

type Duration struct {
	Days   uint
	Months uint
	Years  uint
}

var types []*Type

func SetupSanctionTypes(t []*Type) error {
	res := gokord.DB.Find(&Type{})
	if res.Error != nil {
		return res.Error
	}
	rows, err := res.Rows()
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			utils.SendAlert("sanction.go - closing rows", err.Error())
		}
	}(rows)
	var scannedTypes []*Type
	for rows.Next() {
		var ty *Type
		err = rows.Scan(&ty)
		if err != nil {
			utils.SendAlert("sanction.go - scanning row", err.Error())
		}
		scannedTypes = append(scannedTypes, ty)
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
