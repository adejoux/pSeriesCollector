package config

import "fmt"

/***************************
	HMC backends
	-GetHMCCfgCfgByID(struct)
	-GetHMCCfgMap (map - for interna config use
	-GetHMCCfgArray(Array - for web ui use )
	-AddHMCCfg
	-DelHMCCfg
	-UpdateHMCCfg
  -GetHMCCfgAffectOnDel
***********************************/

/*GetHMCCfgByID get device data by id*/
func (dbc *DatabaseCfg) GetHMCCfgByID(id string) (HMCCfg, error) {
	cfgarray, err := dbc.GetHMCCfgArray("id='" + id + "'")
	if err != nil {
		return HMCCfg{}, err
	}
	if len(cfgarray) > 1 {
		return HMCCfg{}, fmt.Errorf("Error %d results on get HMCCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return HMCCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the HMC config table", id)
	}
	return *cfgarray[0], nil
}

/*GetHMCCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetHMCCfgMap(filter string) (map[string]*HMCCfg, error) {
	cfgarray, err := dbc.GetHMCCfgArray(filter)
	cfgmap := make(map[string]*HMCCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetHMCCfgArray generate an array of devices with all its information */
func (dbc *DatabaseCfg) GetHMCCfgArray(filter string) ([]*HMCCfg, error) {
	var err error
	var devices []*HMCCfg
	//Get Only data for selected devices
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get HMCCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get HMCCfg   data: %v\n", err)
			return nil, err
		}
	}
	return devices, nil
}

/*AddHMCCfg for adding new devices*/
func (dbc *DatabaseCfg) AddHMCCfg(dev HMCCfg) (int64, error) {
	var err error
	var affected int64
	session := dbc.x.NewSession()
	defer session.Close()

	affected, err = session.Insert(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	//no other relation
	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Added new HMC backend Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelHMCCfg for deleting HMC databases from ID*/
func (dbc *DatabaseCfg) DelHMCCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in HMCCfg

	affected, err = session.Where("id='" + id + "'").Delete(&HMCCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully HMC db with ID %s [ %d Devices Affected  ]", id, affecteddev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*UpdateHMCCfg for adding new HMC*/
func (dbc *DatabaseCfg) UpdateHMCCfg(id string, dev HMCCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()

	affected, err = session.Where("id='" + id + "'").UseBool().AllCols().Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Infof("Updated HMC Config Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*GetHMCCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetHMCCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	//	var devices []*HMCCfg
	var obj []*DbObjAction
	/*
		for _, val := range devices {
			obj = append(obj, &DbObjAction{
				Type:     "HMCCfg",
				TypeDesc: "HMC Devices",
				ObID:     val.ID,
				Action:   "Reset HMC Server fro 'default' InfluxDB Server",
			})

		}*/
	return obj, nil
}
