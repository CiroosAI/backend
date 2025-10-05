package models

import "database/sql"

type Setting struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Company        string  `json:"company"`
	Logo           string  `json:"logo"`
	MinWithdraw    float64 `json:"min_withdraw"`
	MaxWithdraw    float64 `json:"max_withdraw"`
	WithdrawCharge float64 `json:"withdraw_charge"`
	Maintenance    bool    `json:"maintenance"`
	ClosedRegister bool    `json:"closed_register"`
	LinkCS         string  `json:"link_cs"`
	LinkGroup      string  `json:"link_group"`
	LinkApp        string  `json:"link_app"`
}

func GetSetting(db *sql.DB) (*Setting, error) {
	setting := &Setting{}
	row := db.QueryRow("SELECT id, name, company, logo, min_withdraw, max_withdraw, withdraw_charge, maintenance, closed_register, link_cs, link_group, link_app FROM settings LIMIT 1")
	err := row.Scan(
		&setting.ID,
		&setting.Name,
		&setting.Company,
		&setting.Logo,
		&setting.MinWithdraw,
		&setting.MaxWithdraw,
		&setting.WithdrawCharge,
		&setting.Maintenance,
		&setting.ClosedRegister,
		&setting.LinkCS,
		&setting.LinkGroup,
		&setting.LinkApp,
	)
	if err != nil {
		return nil, err
	}
	return setting, nil
}
