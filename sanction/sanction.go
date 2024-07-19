package sanction

import (
	"database/sql"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"gorm.io/gorm"
)

type Type struct {
	gorm.Model
	Name        string `gorm:"unique"`
	Description string
	Retention   uint8
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
